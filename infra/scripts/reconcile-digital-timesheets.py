#!/usr/bin/env python3
"""Plan/apply the Digital Timesheet clients using the shared reconciliation engine."""
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
from keycloak_client_provisioning import *  # noqa: F403

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--environment', choices=('production', 'test'), default='production')
    parser.add_argument('--apply', action='store_true')
    parser.add_argument('--backup', type=Path)
    parser.add_argument('--include-test-users', action='store_true')
    args = parser.parse_args()
    if args.apply and not args.backup:
        parser.error('--apply requires a new private --backup file')
    manifest_path = MANIFEST if args.environment == 'production' else MANIFEST.with_name('digital-timesheets-test.json')
    manifest = json.loads(manifest_path.read_text())
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
