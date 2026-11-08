import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

spec = importlib.util.spec_from_file_location(
    'defaults', Path(__file__).with_name('reconcile-keycloak-user-defaults.py'))
defaults = importlib.util.module_from_spec(spec)
spec.loader.exec_module(defaults)


class FakeKeycloak:
    def __init__(self):
        self.writes = []
        self.members = {'already-user'}
        self.roles = [{'id': 'old', 'name': 'utilities_access', 'clientRole': True}]
        self.service_has_user = False

    def get(self, path):
        if path == '/clients':
            return [{'id': 'portal-id', 'clientId': 'portal'},
                    {'id': 'worker-id', 'clientId': 'worker', 'serviceAccountsEnabled': True}]
        if path == '/clients/portal-id/roles/access':
            return {'id': 'access', 'name': 'access', 'clientRole': True, 'containerId': 'portal-id'}
        if path == '/roles/user':
            return {'id': 'user-role', 'name': 'user'}
        if path == '/roles/user/composites':
            return self.roles
        if path == '/clients/worker-id/service-account-user':
            return {'id': 'worker'}
        if path == '/users/worker/role-mappings/realm/composite':
            return [{'id': 'user-role'}] if self.service_has_user else []
        if path.startswith('/users?'):
            return [{'id': 'already-user', 'username': 'human'},
                    {'id': 'new-human', 'username': 'another-human'},
                    {'id': 'worker', 'username': 'service-account-worker'}]
        if path.startswith('/roles/user/users?'):
            return [{'id': value} for value in self.members]
        raise AssertionError('unexpected endpoint: ' + path)

    def request(self, method, path, body):
        self.writes.append((method, path, body))
        if path == '/roles/user/composites':
            if method == 'POST':
                self.roles.extend(body)
            elif method == 'DELETE':
                removed = {role['id'] for role in body}
                self.roles = [role for role in self.roles if role['id'] not in removed]
        elif path == '/users/new-human/role-mappings/realm' and method == 'POST':
            self.members.add('new-human')
        else:
            raise AssertionError('unexpected mutation: ' + path)


class ReconciliationTests(unittest.TestCase):
    def test_dry_run_and_apply_preserve_explicit_roles_and_services(self):
        api = FakeKeycloak()
        plan = defaults.build_plan(api, {'portal': ['access']})
        self.assertEqual(plan['addUserRoleTo'], ['new-human'])
        self.assertEqual(api.writes, [])
        with tempfile.TemporaryDirectory() as directory:
            backup = Path(directory) / 'before.json'
            defaults.apply_plan(api, plan, backup)
            self.assertEqual(backup.stat().st_mode & 0o777, 0o600)
            self.assertEqual(json.loads(backup.read_text())['beforeComposites'][0]['id'], 'old')
            self.assertEqual(defaults.build_plan(api, {'portal': ['access']})['addUserRoleTo'], [])
            with self.assertRaises(FileExistsError):
                defaults.apply_plan(api, plan, backup)
        self.assertEqual(len(api.writes), 3)
        self.assertTrue(all('/role-mappings/clients/' not in path and
                            '/groups' not in path and '/default-roles' not in path
                            for _, path, _ in api.writes))

    def test_refuses_to_change_baseline_inherited_by_service_account(self):
        api = FakeKeycloak()
        api.service_has_user = True
        with self.assertRaisesRegex(ValueError, 'service account'):
            defaults.build_plan(api, {'portal': ['access']})
        self.assertEqual(api.writes, [])

    def test_pagination_reads_all_pages(self):
        class PagedAPI:
            def get(self, path):
                return [{'id': 'one'}, {'id': 'two'}] if 'first=0&' in path else [{'id': 'three'}]
        self.assertEqual([row['id'] for row in defaults.pages(PagedAPI(), '/users', 2)],
                         ['one', 'two', 'three'])


if __name__ == '__main__':
    unittest.main()
