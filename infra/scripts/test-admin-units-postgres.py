#!/usr/bin/env python3
"""Exercise the candidate package against a disposable, read-only PostgreSQL warehouse."""
import json
import os
from pathlib import Path
import secrets
import shutil
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[2]
MODULE_PATH = 'github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units'

def run(*args, **kwargs):
    return subprocess.run(args, check=True, capture_output=True, text=True, **kwargs)

def main():
    name = 'moh-units-test-' + secrets.token_hex(6)
    password = secrets.token_hex(24)
    image = os.environ.get('ADMIN_UNITS_POSTGRES_IMAGE', 'postgres:15-alpine')
    go = os.environ.get('GO_CMD', 'go')
    created = False
    try:
        run('docker', 'run', '--detach', '--rm', '--name', name,
            '-e', 'POSTGRES_PASSWORD=' + password,
            '-p', '127.0.0.1::5432', image)
        created = True
        # The entrypoint's temporary initialization server accepts only socket connections.
        for attempt in range(60):
            ready = subprocess.run(['docker', 'exec', name, 'pg_isready', '-h', '127.0.0.1', '-U', 'postgres'],
                                   capture_output=True)
            if ready.returncode == 0:
                break
            time.sleep(0.5)
        else:
            raise RuntimeError('Disposable PostgreSQL did not become ready')
        sql = """
CREATE SCHEMA dwh;
CREATE TABLE dwh.dim_org_hierarchy(
 dim_org_hierarchy_key bigint, org_unit_id text, org_unit_name text, level text,
 country_uid text, region_uid text, region text, district_uid text, district text,
 sub_county_uid text, sub_county text, division_uid text, division text,
 facility_uid text, facility_name text, row_version bigint, is_current boolean);
INSERT INTO dwh.dim_org_hierarchy
SELECT i, 'unit'||i, NULL, i::text,
 'unit1','unit2','fixture','unit3','fixture','unit5','fixture',
 'unit4','fixture','unit6','fixture',7,true FROM generate_series(1,6) i;
INSERT INTO dwh.dim_org_hierarchy(dim_org_hierarchy_key,level,is_current)
VALUES(-1,'6',true),(99,'6',false);
CREATE ROLE warehouse_reader LOGIN PASSWORD '""" + password + """';
GRANT USAGE ON SCHEMA dwh TO warehouse_reader;
GRANT SELECT ON dwh.dim_org_hierarchy TO warehouse_reader;
ALTER ROLE warehouse_reader SET default_transaction_read_only=on;
"""
        run('docker', 'exec', '-i', name, 'psql', '-U', 'postgres', '-v', 'ON_ERROR_STOP=1', input=sql)
        bindings = json.loads(run('docker', 'inspect', name).stdout)[0]['NetworkSettings']['Ports']['5432/tcp']
        port = bindings[0]['HostPort']
        with tempfile.TemporaryDirectory(prefix='moh-units-postgres-') as tmp:
            root = Path(tmp)
            shutil.copytree(ROOT / 'backend/modules/admin-units', root / 'module')
            consumer = root / 'consumer'
            consumer.mkdir()
            (consumer / 'go.mod').write_text(
                'module example.com/units-postgres-consumer\n\ngo 1.25.4\n\nrequire (\n' +
                MODULE_PATH + ' v0.0.0\n github.com/lib/pq v1.10.9\n)\n\nreplace ' +
                MODULE_PATH + ' => ../module\n')
            sums = [line for line in (ROOT / 'backend/go.sum').read_text().splitlines()
                    if line.startswith('github.com/lib/pq v1.10.9')]
            assert len(sums) == 2, 'Reviewed PostgreSQL driver checksums missing'
            (consumer / 'go.sum').write_text('\n'.join(sums) + '\n')
            shutil.copyfile(root / 'module/testdata/postgres-consumer/main.go', consumer / 'main.go')
            env = {**os.environ, 'GOWORK': 'off', 'GOTOOLCHAIN': 'local', 'GOPROXY': 'off'}
            env['ADMIN_UNITS_TEST_DSN'] = (
                'host=127.0.0.1 port=' + port + ' user=warehouse_reader password=' +
                password + ' dbname=postgres sslmode=disable')
            binary = root / 'consumer-bin'
            run(go, 'build', '-race', '-mod=readonly', '-o', str(binary), '.', cwd=consumer, env=env)
            run(go, 'vet', '-mod=readonly', './...', cwd=consumer, env=env)
            empty = root / 'empty'
            empty.mkdir()
            result = run(str(binary), cwd=empty, env=env)
            print(result.stdout.strip(), flush=True)
    finally:
        if created:
            run('docker', 'rm', '-f', name)

if __name__ == '__main__':
    try:
        main()
    except subprocess.CalledProcessError as error:
        # Never echo command arguments: the fixture DSN contains its password.
        raise RuntimeError('PostgreSQL package verification failed:\n' +
                           (error.stdout or '') + '\n' + (error.stderr or '')) from None
