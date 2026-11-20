#!/usr/bin/env python3
"""Read and verify the backend distribution's published Go module pins."""
import json
import os
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parents[2]
MODULE_PREFIX = 'github.com/mohuganda/moh-sso-dashboard/backend/modules/'
MODULE_VERSION = re.compile(r'v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?')
PIN = re.compile(r'^\s*(?:require\s+)?(' + re.escape(MODULE_PREFIX) + r'\S+)\s+(\S+)', re.MULTILINE)
CHECKSUM = re.compile(r'h1:[A-Za-z0-9+/]{43}=')


def module_pins(backend=ROOT / 'backend'):
    pins = {}
    for module, version in PIN.findall((backend / 'go.mod').read_text()):
        if module in pins or not MODULE_VERSION.fullmatch(version):
            raise ValueError(f'Invalid or duplicate module pin: {module} {version}')
        pins[module] = version
    if not pins:
        raise ValueError('No published backend module pins found')
    return pins


def module_checksums(module, version, backend=ROOT / 'backend'):
    checksums = {}
    wanted = {version, version + '/go.mod'}
    for line in (backend / 'go.sum').read_text().splitlines():
        path, key, checksum = line.split()
        if path == module and key in wanted:
            if key in checksums or not CHECKSUM.fullmatch(checksum):
                raise ValueError(f'Invalid or duplicate checksum for {module}@{key}')
            checksums[key] = checksum
    if checksums.keys() != wanted:
        raise ValueError(f'Missing reviewed checksums for {module}@{version}')
    return checksums


def verify_published_modules(backend=ROOT / 'backend'):
    pins = module_pins(backend)
    expected = {module: module_checksums(module, version, backend) for module, version in pins.items()}
    env = {**os.environ, 'GOWORK': 'off'}
    go = os.environ.get('GO_CMD', 'go')

    def run(*args):
        return subprocess.run([go, *args], cwd=backend, env=env, check=True,
                              text=True, stdout=subprocess.PIPE).stdout

    for module, version in pins.items():
        item = json.loads(run('list', '-mod=readonly', '-m', '-json', module))
        if item.get('Version') != version or item.get('Replace'):
            raise ValueError(f'Unexpected selected module for {module}@{version}: {item}')
        artifact = json.loads(run('mod', 'download', '-json', f'{module}@{version}'))
        if artifact.get('Sum') != expected[module][version] or artifact.get('GoModSum') != expected[module][version + '/go.mod']:
            raise ValueError(f'Published checksums differ for {module}@{version}')
        print(f'Verified {module}@{version}', flush=True)


if __name__ == '__main__':
    verify_published_modules()
