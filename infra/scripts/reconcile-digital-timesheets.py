#!/usr/bin/env python3
"""Plan or apply the Health Workforce Portal clients; dry-run by default."""
import argparse
import copy
import importlib.util
import json
import os
from pathlib import Path

# Reuse the existing administrative transport and its credential-safe error handling.
spec = importlib.util.spec_from_file_location(
    'keycloak_defaults', Path(__file__).with_name('reconcile-keycloak-user-defaults.py'))
transport = importlib.util.module_from_spec(spec)
spec.loader.exec_module(transport)
Keycloak = transport.Keycloak
quote = transport.quote
ROOT = Path(__file__).resolve().parents[2]
MANIFEST = ROOT / 'infra/keycloak/clients/digital-timesheets.json'


def lookup(api, client_id):
    matches = [c for c in api.get('/clients?clientId=' + quote(client_id))
               if c['clientId'] == client_id]
    if len(matches) > 1:
        raise ValueError('Client ID is ambiguous')
    return api.get('/clients/' + quote(matches[0]['id'])) if matches else None


def settings(client, current=None):
    result = {k: copy.deepcopy(v) for k, v in client.items()
              if k not in {'defaultClientScopes', 'optionalClientScopes', 'protocolMappers'}}
    result['attributes'] = {**(current or {}).get('attributes', {}), **result['attributes']}
    # Inherit the actual realm default rather than assuming the PDF's five minutes.
    result['attributes'].pop('access.token.lifespan', None)
    return result


def settings_match(current, desired):
    for key, value in settings(desired).items():
        if key == 'attributes':
            actual = current.get('attributes', {})
            if any(actual.get(k) != v for k, v in value.items()):
                return False
            if actual.get('access.token.lifespan') not in (None, ''):
                return False
        elif current.get(key, False if isinstance(value, bool) else None) != value:
            return False
    return True


def mapper_match(current, desired):
    return all((all(current.get('config', {}).get(name) == value for name, value in v.items())
                if k == 'config' else current.get(k, False if isinstance(v, bool) else None) == v)
               for k, v in desired.items())


def client_state(api, desired):
    current = lookup(api, desired['clientId'])
    if current is None:
        return {'client': None, 'default': [], 'optional': [], 'mappers': []}
    if not current.get('publicClient'):
        raise ValueError('Refusing to convert an existing confidential client')
    path = '/clients/' + quote(current['id'])
    return {'client': current,
            'default': api.get(path + '/default-client-scopes'),
            'optional': api.get(path + '/optional-client-scopes'),
            'mappers': api.get(path + '/protocol-mappers/models')}


def changed(state, desired):
    if state['client'] is None or not settings_match(state['client'], desired):
        return True
    for kind in ('default', 'optional'):
        if {s['name'] for s in state[kind]} != set(desired[kind + 'ClientScopes']):
            return True
    by_name = {m['name']: m for m in state['mappers']}
    return any(not mapper_match(by_name.get(m['name'], {}), m)
               for m in desired['protocolMappers'])


def build_plan(api, manifest, include_test_users=False):
    if api.realm != manifest['realm'] or api.realm != 'moh-realm':
        raise ValueError('The approved target realm is moh-realm; check KEYCLOAK_REALM')
    realm = api.get('')
    if realm['realm'] != api.realm:
        raise ValueError('Keycloak returned a different realm')
    scopes = {s['name']: s['id'] for s in api.get('/client-scopes')}
    missing = {s for c in manifest['clients'] for s in c['defaultClientScopes']} - scopes.keys()
    if missing:
        raise ValueError('Required realm client scopes are missing: ' + ', '.join(sorted(missing)))
    clients = [{'desired': c, 'before': client_state(api, c)} for c in manifest['clients']]
    users = []
    if include_test_users:
        for user in manifest['testUsers']:
            matches = api.get('/users?username=' + quote(user['username']) + '&exact=true')
            if len(matches) == 1:
                matches = [api.get('/users/' + quote(matches[0]['id']))]
            if matches and (len(matches) != 1 or matches[0].get('email', '').lower() != user['email'].lower()):
                raise ValueError('Test username is already owned by another account')
            if not matches:
                users.append(user)
    return {'realm': api.realm, 'realmSettings': {
        k: realm.get(k) for k in ('accessTokenLifespan', 'passwordPolicy', 'bruteForceProtected',
            'failureFactor', 'ssoSessionIdleTimeout', 'ssoSessionMaxLifespan',
            'offlineSessionIdleTimeout', 'registrationAllowed', 'loginWithEmailAllowed',
            'duplicateEmailsAllowed')}, 'smtpConfigured': bool(realm.get('smtpServer')),
        'scopes': scopes, 'clients': clients, 'createTestUsers': users}


def apply_plan(api, plan, backup):
    fd = os.open(backup, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(fd, 'w') as stream:
        json.dump(plan, stream, indent=2)
    for item in plan['clients']:
        desired, before = item['desired'], item['before']
        if before['client'] is None:
            api.request('POST', '/clients', desired)
            before = client_state(api, desired)
        elif not settings_match(before['client'], desired):
            api.request('PUT', '/clients/' + quote(before['client']['id']),
                        settings(desired, before['client']))
        path = '/clients/' + quote(before['client']['id'])
        for kind in ('default', 'optional'):
            assigned = {s['name']: s['id'] for s in before[kind]}
            wanted = set(desired[kind + 'ClientScopes'])
            for name in assigned.keys() - wanted:
                api.request('DELETE', path + '/' + kind + '-client-scopes/' + quote(assigned[name]))
            for name in wanted - assigned.keys():
                api.request('PUT', path + '/' + kind + '-client-scopes/' + quote(plan['scopes'][name]))
        mappers = {m['name']: m for m in before['mappers']}
        for mapper in desired['protocolMappers']:
            existing = mappers.get(mapper['name'])
            if existing is None:
                api.request('POST', path + '/protocol-mappers/models', mapper)
            elif not mapper_match(existing, mapper):
                api.request('PUT', path + '/protocol-mappers/models/' + quote(existing['id']),
                            {**mapper, 'id': existing['id']})
        if changed(client_state(api, desired), desired):
            raise RuntimeError('Client verification failed after applying')
    for user in plan['createTestUsers']:
        api.request('POST', '/users', user)


def summary(plan):
    return {'realm': plan['realm'], 'realmSettings': plan['realmSettings'],
            'smtpConfigured': plan['smtpConfigured'],
            'clientsToCreate': [i['desired']['clientId'] for i in plan['clients'] if i['before']['client'] is None],
            'clientsToUpdate': [i['desired']['clientId'] for i in plan['clients']
                                if i['before']['client'] and changed(i['before'], i['desired'])],
            'testUsersToCreate': len(plan['createTestUsers'])}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--apply', action='store_true')
    parser.add_argument('--backup', type=Path)
    parser.add_argument('--include-test-users', action='store_true')
    args = parser.parse_args()
    if args.apply and not args.backup:
        parser.error('--apply requires a new private --backup file')
    manifest = json.loads(MANIFEST.read_text())
    api = Keycloak(os.environ)
    plan = build_plan(api, manifest, args.include_test_users)
    print(json.dumps(summary(plan)), flush=True)
    if args.apply:
        apply_plan(api, plan, args.backup)
        remaining = summary(build_plan(api, manifest, args.include_test_users))
        if remaining['clientsToCreate'] or remaining['clientsToUpdate'] or remaining['testUsersToCreate']:
            raise RuntimeError('Reconciliation did not converge')
        print('Applied and verified. Test users remain disabled until an administrator sets passwords and enables them.')
    else:
        print('Dry run only; no changes applied.')


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print('Client setup failed: ' + str(error), flush=True)
        raise SystemExit(1)
