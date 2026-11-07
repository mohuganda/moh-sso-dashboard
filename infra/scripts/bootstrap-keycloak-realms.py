#!/usr/bin/env python3
"""Create missing non-production realms on an existing Keycloak; dry-run by default."""
import argparse
import importlib.util
import json
import os
from pathlib import Path
import time
import urllib.error
import urllib.parse
import urllib.request

spec = importlib.util.spec_from_file_location('realm_template', Path(__file__).with_name('prepare-keycloak-realm.py'))
template = importlib.util.module_from_spec(spec)
spec.loader.exec_module(template)
NON_PRODUCTION = {key: value for key, value in template.REALMS.items() if key != 'production'}


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


class ServerAdmin:
    """Operator-only master-realm transport; never reused by portal deployments."""
    def __init__(self, env):
        self.base = env.get('KEYCLOAK_BASE_URL', '').rstrip('/')
        uri = urllib.parse.urlsplit(self.base)
        if (not uri.hostname or uri.username or uri.password or uri.query or uri.fragment
                or any(char.isspace() for char in self.base)
                or (uri.scheme != 'https' and not (uri.scheme == 'http' and uri.hostname in ('localhost', '127.0.0.1', '::1')))):
            raise ValueError('KEYCLOAK_BASE_URL must use HTTPS, or HTTP on loopback for local tests')
        self.client = env.get('KEYCLOAK_BOOTSTRAP_CLIENT_ID', '')
        self.secret = env.get('KEYCLOAK_BOOTSTRAP_CLIENT_SECRET', '')
        self.username = env.get('KEYCLOAK_BOOTSTRAP_USERNAME', '')
        self.password = env.get('KEYCLOAK_BOOTSTRAP_PASSWORD', '')
        client_auth = bool(self.client and self.secret)
        password_auth = bool(self.username and self.password)
        if client_auth == password_auth or (client_auth and (self.username or self.password)) or (password_auth and (self.client or self.secret)):
            raise ValueError('Set either bootstrap client ID/secret or bootstrap username/password, never both')
        self.token = ''
        self.expires = 0
        self.opener = urllib.request.build_opener(NoRedirect())

    def _request(self, method, url, payload=None, headers=None):
        try:
            with self.opener.open(urllib.request.Request(url, data=payload, headers=headers or {}, method=method), timeout=30) as response:
                data = response.read()
                return response.status, json.loads(data) if data else None
        except urllib.error.HTTPError as error:
            # Do not echo Keycloak response bodies, tokens or submitted credentials.
            return error.code, None
        except (OSError, urllib.error.URLError):
            raise RuntimeError('Cannot reach the configured Keycloak server') from None

    def request(self, method, path, body=None):
        # The only allowed writes create one of the three new environment realms.
        if method == 'GET':
            if path not in {'/realms/' + name for name in NON_PRODUCTION.values()}:
                raise ValueError('Only non-production environment realm lookups are allowed')
        elif method != 'POST' or path != '/realms' or not body or body.get('realm') not in NON_PRODUCTION.values():
            raise ValueError('Only creation of non-production environment realms is allowed')
        if time.monotonic() >= self.expires:
            credentials = ({'grant_type': 'client_credentials', 'client_id': self.client, 'client_secret': self.secret}
                           if self.secret else {'grant_type': 'password', 'client_id': 'admin-cli', 'username': self.username, 'password': self.password})
            status, token = self._request('POST', self.base + '/realms/master/protocol/openid-connect/token',
                                          urllib.parse.urlencode(credentials).encode(), {'Content-Type': 'application/x-www-form-urlencoded'})
            if status != 200 or not token or 'access_token' not in token:
                raise RuntimeError('Bootstrap administrator authentication rejected (HTTP %d)' % status)
            self.token = token['access_token']
            self.expires = time.monotonic() + max(1, token.get('expires_in', 60) - 15)
        result = self._request(method, self.base + '/admin' + path,
                               None if body is None else json.dumps(body).encode(),
                               {'Authorization': 'Bearer ' + self.token, 'Content-Type': 'application/json'})
        if method == 'POST' and result[0] in (200, 201, 204):
            # Realm creators may receive new realm-scoped roles during creation.
            self.expires = 0
        return result


def build_plan(api, environments):
    if not isinstance(environments, dict) or not environments or set(environments) - NON_PRODUCTION.keys():
        raise ValueError('Plan may contain only development, test and staging; production is never managed')
    # Render and validate everything before making any server requests.
    definitions = []
    for environment, urls in environments.items():
        if not isinstance(urls, dict) or set(urls) != {'portalUrl', 'apiUrl'} or not all(isinstance(v, str) for v in urls.values()):
            raise ValueError('Each environment requires only portalUrl and apiUrl')
        definitions.append((environment, template.render(environment, urls['portalUrl'], urls['apiUrl'])))
    result = []
    for environment, definition in definitions:
        status, _ = api.request('GET', '/realms/' + definition['realm'])
        if status not in (200, 404):
            raise RuntimeError('Cannot inspect environment realm (HTTP %d); no changes applied' % status)
        result.append({'environment': environment, 'realm': definition['realm'],
                       'action': 'skip-existing' if status == 200 else 'create', 'definition': definition})
    return result


def summary(plan):
    return [{key: item[key] for key in ('environment', 'realm', 'action')} for item in plan]


def apply_plan(api, plan, record_path):
    # Never accept a modified/foreign plan that could write to production.
    for item in plan:
        if (item.get('environment') not in NON_PRODUCTION
                or item.get('realm') != NON_PRODUCTION[item['environment']]
                or item.get('definition', {}).get('realm') != item['realm']
                or item.get('action') not in ('create', 'skip-existing')):
            raise ValueError('Invalid non-production realm creation plan')
    fd = os.open(record_path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    record = {'server': api.base, 'realms': [{**row, 'status': 'pending' if row['action'] == 'create' else 'untouched'} for row in summary(plan)]}
    with os.fdopen(fd, 'w') as stream:
        def save():
            stream.seek(0)
            json.dump(record, stream, indent=2)
            stream.truncate()
            stream.flush()
            os.fsync(stream.fileno())
        save()
        for item, row in zip(plan, record['realms']):
            if item['action'] == 'skip-existing':
                continue
            # Recheck after review so a concurrently created realm is not replaced.
            status, _ = api.request('GET', '/realms/' + item['realm'])
            if status == 200:
                row['status'] = 'untouched-existing'
            elif status == 404:
                row['status'] = 'creation-started'
                save()
                status, _ = api.request('POST', '/realms', item['definition'])
                if status == 409:
                    row['status'] = 'untouched-concurrent-create'
                elif status in (200, 201, 204):
                    # Record creation before verifying; a later read failure must not lose it.
                    row['status'] = 'created'
                    save()
                    verify, _ = api.request('GET', '/realms/' + item['realm'])
                    if verify != 200:
                        raise RuntimeError('Realm created but verification failed (HTTP %d); inspect the private record' % verify)
                    row['status'] = 'created-verified'
                else:
                    row['status'] = 'creation-rejected'
                    save()
                    raise RuntimeError('Realm creation rejected (HTTP %d); inspect the private record' % status)
            else:
                raise RuntimeError('Realm recheck failed (HTTP %d); inspect the private record' % status)
            save()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--plan', required=True, type=Path, help='Environment portal/API URLs; no credentials')
    parser.add_argument('--apply', action='store_true')
    parser.add_argument('--record', type=Path, help='New private creation record; required for apply')
    args = parser.parse_args()
    if args.apply and not args.record:
        parser.error('--apply requires a new --record file')
    environments = json.loads(args.plan.read_text())
    api = ServerAdmin(os.environ)
    plan = build_plan(api, environments)
    print(json.dumps({'server': api.base, 'realms': summary(plan)}), flush=True)
    if args.apply:
        apply_plan(api, plan, args.record)
        print('Missing realms created and verified. Existing realms and production were untouched.')
    else:
        print('Dry run only. Existing realms will be skipped; production is never managed.')


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        # Unexpected parser/filesystem failures should not echo credential-bearing inputs.
        if isinstance(error, (ValueError, RuntimeError)) and not isinstance(error, json.JSONDecodeError):
            print('Realm setup failed: ' + str(error))
        else:
            print('Realm setup failed: check the plan and private record paths')
        raise SystemExit(1)
