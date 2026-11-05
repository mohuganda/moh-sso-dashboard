#!/usr/bin/env python3
"""Reconcile the human-user baseline without modifying explicit access grants.

Dry-run by default. Credentials come from the existing KEYCLOAK_* environment.
Generate --policy with `moh-sso system-rbac user-defaults`. No credentials or
user identities are printed. Realm-wide default roles are deliberately untouched.
"""
import argparse
import json
import os
from pathlib import Path
import time
import urllib.error
import urllib.parse
import urllib.request


class Keycloak:
    def __init__(self, env):
        self.base = (env.get('KEYCLOAK_INTERNAL_URL') or env.get('KEYCLOAK_BASE_URL', '')).rstrip('/')
        self.realm = env.get('KEYCLOAK_REALM', '')
        self.client = env.get('KEYCLOAK_ADMIN_CLIENT_ID', '')
        self.secret = env.get('KEYCLOAK_ADMIN_CLIENT_SECRET', '')
        if not all([self.base, self.realm, self.client, self.secret]):
            raise ValueError('Keycloak admin connection environment is incomplete')
        self.admin = self.base + '/admin/realms/' + quote(self.realm)
        self.token = ''
        self.expires = 0

    def request(self, method, path, body=None):
        if time.monotonic() >= self.expires:
            payload = urllib.parse.urlencode({
                'grant_type': 'client_credentials', 'client_id': self.client,
                'client_secret': self.secret,
            }).encode()
            token = self._request('POST', self.base + '/realms/' + quote(self.realm) +
                                  '/protocol/openid-connect/token', payload,
                                  {'Content-Type': 'application/x-www-form-urlencoded'})
            self.token = token['access_token']
            self.expires = time.monotonic() + max(1, token.get('expires_in', 60) - 15)
        payload = None if body is None else json.dumps(body).encode()
        return self._request(method, self.admin + path, payload, {
            'Authorization': 'Bearer ' + self.token, 'Content-Type': 'application/json',
        })

    @staticmethod
    def _request(method, url, payload, headers):
        try:
            with urllib.request.urlopen(urllib.request.Request(
                    url, data=payload, headers=headers, method=method), timeout=30) as response:
                data = response.read()
                return json.loads(data) if data else None
        except urllib.error.HTTPError as error:
            # Keycloak error bodies can include request data; report status only.
            raise RuntimeError('Keycloak request rejected (HTTP %d)' % error.code) from None

    def get(self, path):
        return self.request('GET', path)


def quote(value):
    return urllib.parse.quote(value, safe='')


def pages(api, path, size=200):
    first = 0
    while True:
        batch = api.get(path + ('&' if '?' in path else '?') +
                        'first=%d&max=%d' % (first, size))
        yield from batch
        if len(batch) < size:
            break
        first += len(batch)


def role_payload(role):
    return {key: role[key] for key in ('id', 'name', 'clientRole', 'containerId') if key in role}


def build_plan(api, policy):
    if not isinstance(policy, dict) or not policy or any(
            not isinstance(client, str) or not isinstance(roles, list) or not roles or
            any(not isinstance(role, str) or not role.strip() for role in roles)
            for client, roles in policy.items()):
        raise ValueError('Policy must map client IDs to nonempty role-name lists')
    clients = api.get('/clients')
    by_name = {client['clientId']: client for client in clients}
    desired = []
    for client_id, names in sorted(policy.items()):
        if client_id not in by_name:
            raise ValueError('Required baseline client is missing: ' + client_id)
        client = by_name[client_id]
        for name in names:
            role = api.get('/clients/' + quote(client['id']) + '/roles/' + quote(name))
            if role.get('composite'):
                raise ValueError('Baseline roles must not inherit unreviewed composite roles')
            desired.append(role_payload(role))
    user_role = role_payload(api.get('/roles/user'))
    current = api.get('/roles/user/composites')
    # Explicitly resolve service-account IDs: /users omits them on some versions.
    service_ids = set()
    for client in clients:
        if client.get('serviceAccountsEnabled'):
            account = api.get('/clients/' + quote(client['id']) + '/service-account-user')
            service_ids.add(account['id'])
            roles = api.get('/users/' + quote(account['id']) + '/role-mappings/realm/composite')
            if any(role['id'] == user_role['id'] for role in roles):
                raise ValueError('A service account inherits user; resolve it before changing defaults')
    human_ids = {
        user['id'] for user in pages(api, '/users?briefRepresentation=false')
        if user['id'] not in service_ids and not user.get('serviceAccountClientId') and
        not user.get('username', '').startswith('service-account-')
    }
    assigned_ids = {user['id'] for user in pages(api, '/roles/user/users')}
    current_ids = {role['id'] for role in current}
    desired_ids = {role['id'] for role in desired}
    return {
        'userRole': user_role,
        'beforeComposites': [role_payload(role) for role in current],
        'desiredComposites': desired,
        'addComposites': [role for role in desired if role['id'] not in current_ids],
        'removeComposites': [role_payload(role) for role in current if role['id'] not in desired_ids],
        'addUserRoleTo': sorted(human_ids - assigned_ids),
        'humanIds': sorted(human_ids),
        'humanCount': len(human_ids), 'serviceAccountCount': len(service_ids),
    }


def summary(plan):
    return {
        'human_accounts': plan['humanCount'],
        'service_accounts_excluded': plan['serviceAccountCount'],
        'default_roles_to_add': len(plan['addComposites']),
        'default_roles_to_remove': len(plan['removeComposites']),
        'human_accounts_to_add_user_role': len(plan['addUserRoleTo']),
    }


def apply_plan(api, plan, backup):
    # Exclusive creation prevents overwriting a previous recovery snapshot.
    fd = os.open(backup, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(fd, 'w') as stream:
        json.dump(plan, stream, indent=2)
        stream.flush()
        os.fsync(stream.fileno())
    if plan['addComposites']:
        api.request('POST', '/roles/user/composites', plan['addComposites'])
    if plan['removeComposites']:
        api.request('DELETE', '/roles/user/composites', plan['removeComposites'])
    actual = {role['id'] for role in api.get('/roles/user/composites')}
    if actual != {role['id'] for role in plan['desiredComposites']}:
        raise RuntimeError('User-role composites differ after applying; use the private backup to recover')
    for index, user_id in enumerate(plan['addUserRoleTo'], 1):
        api.request('POST', '/users/' + quote(user_id) + '/role-mappings/realm', [plan['userRole']])
        if index % 200 == 0:
            print(json.dumps({'human_accounts_processed': index}), flush=True)
    assigned = {user['id'] for user in pages(api, '/roles/user/users')}
    if not set(plan['humanIds']).issubset(assigned):
        raise RuntimeError('Not all human user-role assignments could be verified')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--policy', required=True, type=Path)
    parser.add_argument('--apply', action='store_true')
    parser.add_argument('--backup', type=Path, help='Required new private snapshot file when applying')
    args = parser.parse_args()
    if args.apply and not args.backup:
        parser.error('--apply requires --backup')
    api = Keycloak(os.environ)
    plan = build_plan(api, json.loads(args.policy.read_text()))
    print(json.dumps(summary(plan)), flush=True)
    if args.apply:
        apply_plan(api, plan, args.backup)
        print(json.dumps({'applied': True, 'human_user_defaults_verified': True}), flush=True)
    else:
        print('Dry run only; no changes applied.')


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        # Avoid a traceback containing request/connection objects with secrets.
        print('Reconciliation failed: ' + str(error), flush=True)
        raise SystemExit(1)
