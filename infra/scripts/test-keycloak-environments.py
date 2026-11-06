#!/usr/bin/env python3
import importlib.util
import json
from pathlib import Path
import sys
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parent))

def module(name, filename):
    spec = importlib.util.spec_from_file_location(name, Path(__file__).with_name(filename))
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result

bootstrap = module('bootstrap', 'prepare-keycloak-realm.py')
clients = module('clients', 'reconcile-keycloak-clients.py')

class Environments(unittest.TestCase):
    def test_realms_exclude_humans_credentials_and_external_app_urls(self):
        names = set()
        for environment, name in bootstrap.REALMS.items():
            realm = bootstrap.render(environment, 'https://portal.example.org/portal', 'https://portal.example.org/api/v1')
            names.add(realm['realm'])
            self.assertEqual(realm['realm'], name)
            self.assertTrue(all(u.get('serviceAccountClientId') for u in realm['users']))
            encoded = json.dumps(realm)
            for forbidden in ('"credentials"', '"secret"', '"privateKey"', 'timesheet.health.go.ug', 'dashboards.health.go.ug', 'localhost:3000'):
                self.assertNotIn(forbidden, encoded)
            web = next(c for c in realm['clients'] if c['clientId'] == 'dashboard-web')
            self.assertEqual(web['redirectUris'], ['https://portal.example.org/api/v1/auth/callback'])
        self.assertEqual(len(names), 4)
        with self.assertRaises(ValueError):
            bootstrap.render('production', 'http://localhost:3000', 'http://localhost:9000/api/v1')

    def test_profiles_and_promotion_require_target_environment_review(self):
        spec = {'version': 1, 'clientId': 'fixture-app', 'name': 'Fixture', 'owner': 'Team',
                'supportEmail': 'support@example.invalid', 'profile': 'public-web', 'environment': 'staging',
                'rootUrl': 'https://app.example.org', 'redirectUris': ['https://app.example.org/callback'],
                'logoutUris': ['https://app.example.org/login'], 'webOrigins': ['https://app.example.org']}
        client = clients.render(spec, 'staging')
        self.assertFalse(client['enabled'])
        self.assertEqual(client['attributes']['pkce.code.challenge.method'], 'S256')
        with self.assertRaises(ValueError):
            clients.render(spec, 'production')
        for bad in ('https://app.example.org/*', 'http://app.example.org/callback'):
            with self.assertRaises(ValueError):
                clients.render({**spec, 'redirectUris': [bad]}, 'staging')
        service = clients.render({**spec, 'profile': 'service', 'rootUrl': '', 'redirectUris': [], 'logoutUris': [], 'webOrigins': []}, 'staging')
        self.assertTrue(service['serviceAccountsEnabled'])
        self.assertFalse(service['standardFlowEnabled'])
        self.assertEqual(service['defaultClientScopes'], [])

if __name__ == '__main__':
    unittest.main()
