#!/usr/bin/env python3
"""Provision a reviewed onboarding document in one environment; dry-run by default."""
import argparse
import copy
import json
import os
from pathlib import Path
import re
import urllib.parse
from keycloak_client_provisioning import Keycloak, ROOT, apply_plan, build_plan, summary

REALMS = {'development': 'moh-dev', 'test': 'moh-test', 'staging': 'moh-staging', 'production': 'moh-realm'}


def render(spec, environment):
    if spec.get('version') != 1 or spec.get('environment') != environment:
        raise ValueError('Review the document version and target environment')
    if not re.fullmatch(r'[a-z0-9][a-z0-9-]{1,99}', spec.get('clientId', '')):
        raise ValueError('Invalid client ID')
    if not all(spec.get(k, '').strip() for k in ('name', 'owner', 'supportEmail')) or '@' not in spec['supportEmail']:
        raise ValueError('Application name, owner and support email are required')
    profiles = json.loads((ROOT / 'backend/internal/keycloak/client-profiles.json').read_text())
    matches = [p for p in profiles if p['id'] == spec.get('profile')]
    if not matches:
        raise ValueError('Unknown client profile')
    client = copy.deepcopy(matches[0]['settings'])
    service = spec['profile'] == 'service'
    if not service and not all(spec.get(k) for k in ('rootUrl', 'redirectUris', 'logoutUris', 'webOrigins')):
        raise ValueError('Exact callback, logout, root and origin URLs are required')
    if service and (spec.get('launcher') or any(spec.get(k) for k in ('rootUrl', 'redirectUris', 'logoutUris', 'webOrigins'))):
        raise ValueError('Service integrations do not use browser URLs or launcher entries')
    for key in ('rootUrl', 'redirectUris', 'logoutUris', 'webOrigins'):
        values = [spec[key]] if key == 'rootUrl' and spec.get(key) else spec.get(key, [])
        for raw in values:
            url = urllib.parse.urlsplit(raw)
            if not url.hostname or url.username or url.password or url.fragment or re.search(r'[*\\\s]', raw) or url.scheme not in ('http', 'https'):
                raise ValueError('Use exact absolute HTTP(S) URLs without credentials or fragments')
            if url.scheme == 'http' and (environment not in ('development', 'test') or url.hostname not in ('localhost', '127.0.0.1', '::1')):
                raise ValueError('HTTP is limited to loopback in development/test')
            if key == 'webOrigins' and (url.path or url.query):
                raise ValueError('Web origins cannot contain a path or query')
    client.update(clientId=spec['clientId'], name=spec['name'], description=spec.get('description', ''),
                  rootUrl=spec.get('rootUrl', ''), baseUrl=spec.get('rootUrl', ''),
                  redirectUris=spec.get('redirectUris', []), webOrigins=spec.get('webOrigins', []))
    attrs = client['attributes']
    attrs.update({'portal.onboarding.profile': spec['profile'], 'portal.onboarding.environment': environment,
                  'portal.owner': spec['owner'], 'portal.supportEmail': spec['supportEmail'], 'ui.launchUrl': '', 'ui.home': ''})
    if spec.get('logoutUris'):
        attrs['post.logout.redirect.uris'] = '##'.join(spec['logoutUris'])
    if spec.get('launcher'):
        attrs.update({'portal.system': 'true', 'ui.systemType': 'external', 'ui.launchMode': 'new_tab',
                      'ui.launchUrl': spec['rootUrl'], 'ui.displayInLauncher': 'true'})
    return client


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--spec', required=True, type=Path)
    parser.add_argument('--environment', required=True, choices=REALMS)
    parser.add_argument('--apply', action='store_true')
    parser.add_argument('--backup', type=Path)
    args = parser.parse_args()
    if args.apply and not args.backup:
        parser.error('--apply requires a new private --backup file')
    api = Keycloak(os.environ)
    if api.realm != REALMS[args.environment]:
        raise ValueError('KEYCLOAK_REALM does not match the target environment')
    client = render(json.loads(args.spec.read_text()), args.environment)
    plan = build_plan(api, {'realm': api.realm, 'clients': [client]})
    before = plan['clients'][0]['before']['client']
    if before and (before.get('enabled') or before.get('attributes', {}).get('portal.onboarding.profile') != client['attributes']['portal.onboarding.profile']):
        raise ValueError('Existing client must be disabled and owned by the same onboarding profile')
    print(json.dumps(summary(plan)), flush=True)
    if args.apply:
        apply_plan(api, plan, args.backup)
        print('Applied and verified; keep disabled until integration checks and access mapping are complete.')
    else:
        print('Dry run only; no changes applied.')


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print('Client setup failed: ' + str(error))
        raise SystemExit(1)
