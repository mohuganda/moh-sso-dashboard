#!/usr/bin/env python3
"""Create and validate an immutable image-pair deployment bundle (stdlib only)."""
import argparse
import hashlib
import json
from pathlib import Path, PurePosixPath
import re
import tarfile

ROOT = Path(__file__).resolve().parents[2]
HEX = re.compile(r'[0-9a-f]{40}')
DIGEST = re.compile(r'sha256:[0-9a-f]{64}')
TAG = re.compile(r'(?:[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?|sha-[0-9a-f]{7,40})')
VERSION = re.compile(r'[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?')
REPOSITORY = re.compile(r'ghcr\.io/[a-z0-9_.-]+/[a-z0-9_.-]+')


def validate_manifest(data):
    if data.get('schema_version') != 1 or not HEX.fullmatch(data.get('source_commit', '')):
        raise ValueError('Invalid manifest schema/source commit')
    if not re.fullmatch(r'[0-9a-f]{40}-[0-9]+-[0-9]+', data.get('release_id', '')):
        raise ValueError('Invalid release ID')
    if not data['release_id'].startswith(data['source_commit'] + '-'):
        raise ValueError('Release ID does not match its source commit')
    if set(data.get('images', {})) != {'backend', 'frontend'}:
        raise ValueError('Both backend and frontend image metadata are required')
    for component in ('backend', 'frontend'):
        image = data['images'][component]
        for key, pattern in [('repository', REPOSITORY), ('digest', DIGEST), ('tag', TAG), ('version', VERSION), ('commit', HEX)]:
            if not pattern.fullmatch(image.get(key, '')):
                raise ValueError(f'Invalid {component} {key}')
        if image['commit'] != data['source_commit']:
            raise ValueError('Image/source commits must match for a paired deployment')
        if not image['repository'].endswith('-' + component):
            raise ValueError('Image repository/component mismatch')
    return data


ENV_NAME = re.compile(r'[A-Z_][A-Z0-9_]*')
SENSITIVE_NAME = re.compile(r'PASSWORD|SECRET|PRIVATE_KEY|API_KEY|SYMMETRIC_KEY|CREDENTIAL')


def validate_environment_contract(data):
    if not isinstance(data, dict) or data.get('schema_version') != 1 or not isinstance(data.get('settings'), list):
        raise ValueError('Invalid release environment contract')
    seen = set()
    for setting in data['settings']:
        if not isinstance(setting, dict):
            raise ValueError('Invalid release environment setting')
        name = setting.get('name', '')
        default = setting.get('default', '')
        alternatives = setting.get('alternatives', [])
        if (not isinstance(name, str) or not ENV_NAME.fullmatch(name) or name in seen or
                not isinstance(default, str) or not isinstance(setting.get('required'), bool) or
                not isinstance(alternatives, list) or
                any(not isinstance(key, str) or not ENV_NAME.fullmatch(key) for key in alternatives)):
            raise ValueError('Invalid release environment setting')
        if default and (SENSITIVE_NAME.search(name) or not re.fullmatch(r'[A-Za-z0-9_./:-]{1,128}', default)):
            raise ValueError('Unsafe default for environment setting: ' + name)
        seen.add(name)
    return data


def build_environment_contract(template, policy):
    # Read names only: example values can be placeholders or local operator edits.
    names = re.findall(r'^([A-Z_][A-Z0-9_]*)=', template, re.MULTILINE)
    if len(names) != len(set(names)):
        raise ValueError('Duplicate setting names in production environment template')
    if policy.get('schema_version') != 1:
        raise ValueError('Unsupported deployment environment policy')
    defaults, required, alternatives = policy.get('defaults', {}), policy.get('required', []), policy.get('alternatives', {})
    if (set(defaults) | set(required) | set(alternatives)) - set(names):
        raise ValueError('Deployment environment policy references unknown setting names')
    return validate_environment_contract({'schema_version': 1, 'settings': [
        {'name': name, 'default': defaults.get(name, ''), 'required': name in required,
         'alternatives': alternatives.get(name, [])} for name in names]})


def allowed(name):
    path = PurePosixPath(name)
    if path.is_absolute() or '..' in path.parts:
        return False
    return (name in {'deployment.json', 'environment-contract.json'} or
            name.startswith(('infra/compose/', 'infra/nginx/')) or
            name in {'infra/scripts/compose.sh', 'infra/scripts/cd_server.py',
                     'infra/scripts/cd_bundle.py', 'infra/scripts/verify-frontend-deployment.mjs'})


def package(metadata_dir, output, commit, run_id, attempt):
    if not HEX.fullmatch(commit) or not str(run_id).isdigit() or not str(attempt).isdigit():
        raise ValueError('Invalid build identity')
    data = {'schema_version': 1, 'release_id': f'{commit}-{run_id}-{attempt}',
            'source_commit': commit, 'images': {}}
    for component in ('backend', 'frontend'):
        data['images'][component] = json.loads((metadata_dir / f'{component}.json').read_text())
    validate_manifest(data)
    output.mkdir(parents=True, exist_ok=True)
    manifest = output / 'deployment.json'
    manifest.write_text(json.dumps(data, indent=2) + '\n')
    contract = output / 'environment-contract.json'
    contract.write_text(json.dumps(build_environment_contract(
        (ROOT / 'infra/env/prod.env.example').read_text(),
        json.loads((ROOT / 'infra/env/deployment-policy.json').read_text())), indent=2) + '\n')
    archive = output / 'deployment.tar.gz'
    files = sorted(p for p in (ROOT / 'infra').rglob('*') if p.is_file() and allowed(p.relative_to(ROOT).as_posix()))
    with tarfile.open(archive, 'w:gz') as tar:
        tar.add(manifest, arcname='deployment.json', recursive=False)
        tar.add(contract, arcname='environment-contract.json', recursive=False)
        for path in files:
            tar.add(path, arcname=path.relative_to(ROOT).as_posix(), recursive=False)
    (output / 'deployment.tar.gz.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest() + '  deployment.tar.gz\n')
    return data


def extract(archive, checksum, destination):
    if destination.is_symlink():
        raise ValueError('Release destination cannot be a symlink')
    destination = destination.parent.resolve() / destination.name
    expected = checksum.read_text().split()[0]
    if not re.fullmatch(r'[0-9a-f]{64}', expected) or hashlib.sha256(archive.read_bytes()).hexdigest() != expected:
        raise ValueError('Deployment archive checksum mismatch')
    with tarfile.open(archive, 'r:gz') as tar:
        entries = tar.getmembers()
        names = [entry.name for entry in entries]
        if len(names) != len(set(names)) or len(names) > 1000:
            raise ValueError('Duplicate/excessive archive entries')
        if sum(entry.size for entry in entries) > 20 * 1024 * 1024:
            raise ValueError('Deployment archive is too large')
        for entry in entries:
            if not entry.isfile() or not allowed(entry.name):
                raise ValueError('Unsafe or unexpected deployment archive entry')
        if 'deployment.json' not in names:
            raise ValueError('Deployment manifest missing')
        data = validate_manifest(json.load(tar.extractfile('deployment.json')))
        for required in ('environment-contract.json', 'infra/scripts/cd_server.py', 'infra/scripts/compose.sh',
                         'infra/scripts/verify-frontend-deployment.mjs', 'infra/compose/docker-compose.yml',
                         'infra/compose/docker-compose.prod.yml', 'infra/compose/docker-compose-nginx.yml'):
            if required not in names:
                raise ValueError('Incomplete deployment infrastructure bundle')
        validate_environment_contract(json.load(tar.extractfile('environment-contract.json')))
        if destination.exists():
            actual = {p.relative_to(destination).as_posix() for p in destination.rglob('*') if p.is_file()}
            if actual != set(names):
                raise ValueError('Existing release has different files')
            for entry in entries:
                target = destination / entry.name
                if target.is_symlink() or target.resolve() != target.absolute() or target.read_bytes() != tar.extractfile(entry).read():
                    raise ValueError('Existing release differs from the approved artifact')
            return data
        destination.mkdir(parents=True, mode=0o750)
        for entry in entries:
            target = destination / entry.name
            target.parent.mkdir(parents=True, exist_ok=True)
            with tar.extractfile(entry) as source, target.open('xb') as dest:
                dest.write(source.read())
            target.chmod(0o640)
    return data


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest='action', required=True)
    build = sub.add_parser('package')
    build.add_argument('--metadata-dir', type=Path, required=True)
    build.add_argument('--output', type=Path, required=True)
    build.add_argument('--commit', required=True)
    build.add_argument('--run-id', required=True)
    build.add_argument('--attempt', default='1')
    verify = sub.add_parser('extract')
    verify.add_argument('--archive', type=Path, required=True)
    verify.add_argument('--checksum', type=Path, required=True)
    verify.add_argument('--destination', type=Path, required=True)
    args = parser.parse_args()
    if args.action == 'package':
        data = package(args.metadata_dir, args.output, args.commit, args.run_id, args.attempt)
    else:
        data = extract(args.archive, args.checksum, args.destination)
    print(json.dumps(data, indent=2))


if __name__ == '__main__':
    main()
