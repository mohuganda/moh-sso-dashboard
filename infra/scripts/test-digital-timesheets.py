#!/usr/bin/env python3
"""Client profile and reconciliation contracts (no live realm access)."""
import copy
import importlib.util
import json
from pathlib import Path
import stat
import tempfile
import unittest

spec = importlib.util.spec_from_file_location('timesheets', Path(__file__).with_name('reconcile-digital-timesheets.py'))
setup = importlib.util.module_from_spec(spec)
spec.loader.exec_module(setup)
MANIFEST = json.loads(setup.MANIFEST.read_text())
TEST_USERS = json.loads(setup.MANIFEST.with_name('digital-timesheets-test.json').read_text())['testUsers']


class FakeKeycloak:
    realm = 'moh-realm'

    def __init__(self):
        self.clients = {}
        self.users = {}
        self.writes = []
        self.scopes = [{'id': name, 'name': name} for name in ('profile', 'email', 'roles', 'offline_access')]

    def get(self, path):
        if path == '':
            return {'realm': self.realm, 'accessTokenLifespan': 300, 'smtpServer': {}}
        if path == '/client-scopes':
            return self.scopes
        if path.startswith('/clients?clientId='):
            return [self.clients[path.split('=')[1]]['client']] if path.split('=')[1] in self.clients else []
        if path.startswith('/users?username='):
            name = path.split('=')[1].split('&')[0]
            return [self.users[name]] if name in self.users else []
        if path.startswith('/users/'):
            return self.users[path.split('/')[2]]
        parts = path.split('/')
        item = self.clients[parts[2]]
        if len(parts) == 3:
            return copy.deepcopy(item['client'])
        if parts[3] in ('default-client-scopes', 'optional-client-scopes'):
            return copy.deepcopy(item[parts[3].split('-')[0]])
        if parts[3:] == ['protocol-mappers', 'models']:
            return copy.deepcopy(item['mappers'])
        raise AssertionError(path)

    def request(self, method, path, body=None):
        self.writes.append((method, path, body))
        if method == 'POST' and path == '/clients':
            cid = body['clientId']
            self.clients[cid] = {'client': {**setup.settings(body), 'id': cid},
                'default': self.scopes[:3], 'optional': self.scopes[3:],
                'mappers': [{**m, 'id': m['name']} for m in body['protocolMappers']]}
        elif method == 'POST' and path == '/users':
            self.users[body['username']] = {**copy.deepcopy(body), 'id': body['username']}
        else:
            parts = path.split('/')
            item = self.clients[parts[2]]
            if method == 'PUT' and len(parts) == 3:
                item['client'] = {**body, 'id': parts[2]}
            elif 'client-scopes' in parts[3]:
                kind = parts[3].split('-')[0]
                if method == 'DELETE':
                    item[kind] = [s for s in item[kind] if s['id'] != parts[4]]
                else:
                    item[kind].append(next(s for s in self.scopes if s['id'] == parts[4]))
            elif parts[3:5] == ['protocol-mappers', 'models']:
                item['mappers'] = [m for m in item['mappers'] if m['name'] != body['name']]
                item['mappers'].append({**body, 'id': body['name']})
            else:
                raise AssertionError((method, path))


class ClientContracts(unittest.TestCase):
    def test_exports_and_exact_client_security_profile(self):
        for filename in ('realm-export.json', 'realm-export-prod.json'):
            realm = json.loads((setup.ROOT / 'infra/keycloak' / filename).read_text())
            for desired in MANIFEST['clients']:
                client = next(c for c in realm['clients'] if c['clientId'] == desired['clientId'])
                self.assertEqual(client, desired)
                self.assertTrue(client['publicClient'])
                self.assertTrue(client['standardFlowEnabled'])
                for key in ('implicitFlowEnabled', 'directAccessGrantsEnabled', 'serviceAccountsEnabled', 'fullScopeAllowed', 'frontchannelLogout'):
                    self.assertFalse(client[key], key)
                self.assertNotIn('secret', client)
                self.assertEqual(client['attributes']['pkce.code.challenge.method'], 'S256')
                self.assertEqual(client['redirectUris'], [client['rootUrl'] + '/auth/sso'])
                self.assertEqual(client['attributes']['post.logout.redirect.uris'], client['rootUrl'] + '/login')
                self.assertEqual(client['webOrigins'], [client['rootUrl']])
                self.assertEqual(client['optionalClientScopes'], [])
                self.assertEqual(client['attributes']['portal.system'], 'false')
                self.assertNotIn('access.token.lifespan', client['attributes'])
                claims = {m['config'].get('claim.name') for m in client['protocolMappers']
                          if m['config'].get('access.token.claim') == 'true'}
                self.assertTrue({'email', 'email_verified'} <= claims)
                self.assertIn('oidc-sub-mapper', {m['protocolMapper'] for m in client['protocolMappers']})

    def test_dry_run_apply_idempotence_and_disabled_test_accounts(self):
        api = FakeKeycloak()
        plan = setup.build_plan(api, {**MANIFEST, "testUsers": TEST_USERS}, True)
        self.assertEqual(api.writes, [])
        self.assertEqual(setup.summary(plan)['testUsersToCreate'], 4)
        with tempfile.TemporaryDirectory() as tmp:
            backup = Path(tmp) / 'before.json'
            setup.apply_plan(api, plan, backup)
            self.assertEqual(stat.S_IMODE(backup.stat().st_mode), 0o600)
            with self.assertRaises(FileExistsError):
                setup.apply_plan(api, plan, backup)
            api.writes.clear()
            plan = setup.build_plan(api, {**MANIFEST, "testUsers": TEST_USERS}, True)
            self.assertEqual(setup.summary(plan)['clientsToUpdate'], [])
            setup.apply_plan(api, plan, Path(tmp) / 'repeat.json')
            self.assertEqual(api.writes, [])
        self.assertEqual(sum(u['emailVerified'] for u in api.users.values()), 3)
        self.assertTrue(all(not u['enabled'] and 'credentials' not in u for u in api.users.values()))
        for client in api.clients.values():
            self.assertEqual({s['name'] for s in client['default']}, {'profile', 'email'})
            self.assertEqual(client['optional'], [])

    def test_wrong_realm_and_existing_identity_collision_are_rejected(self):
        api = FakeKeycloak()
        api.realm = 'MoH'
        with self.assertRaises(ValueError):
            setup.build_plan(api, MANIFEST)
        api.realm = 'moh-realm'
        name = TEST_USERS[0]['username']
        api.users[name] = {'id': name, 'email': 'someone@example.invalid'}
        with self.assertRaises(ValueError):
            setup.build_plan(api, {**MANIFEST, "testUsers": TEST_USERS}, True)
        self.assertEqual(api.writes, [])

    def test_reconcile_insecure_settings_without_overwriting_unrelated_attributes(self):
        api = FakeKeycloak()
        with tempfile.TemporaryDirectory() as tmp:
            setup.apply_plan(api, setup.build_plan(api, MANIFEST), Path(tmp) / 'create.json')
            item = api.clients['digital-timesheets']
            item['client']['attributes']['operator.note'] = 'keep'
            item['client']['attributes']['access.token.lifespan'] = '999'
            item['client']['implicitFlowEnabled'] = True
            item['mappers'][0]['config'] = {'access.token.claim': 'false'}
            setup.apply_plan(api, setup.build_plan(api, MANIFEST), Path(tmp) / 'repair.json')
        actual = api.clients['digital-timesheets']['client']
        self.assertEqual(actual['attributes']['operator.note'], 'keep')
        self.assertNotIn('access.token.lifespan', actual['attributes'])
        self.assertFalse(actual['implicitFlowEnabled'])


if __name__ == '__main__':
    unittest.main()
