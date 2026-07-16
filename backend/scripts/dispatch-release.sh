#!/bin/sh

set -eu

usage() {
  echo "usage: $0 <version> [repo]" >&2
  echo "example: $0 1.2.3 mohuganda/moh-sso-dashboard" >&2
}

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  usage
  exit 2
fi

version=${1#backend/v}
version=${version#v}
repo=${2:-mohuganda/moh-sso-dashboard}

if ! printf '%s' "$version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'; then
  echo "invalid backend semantic version: $version" >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "cannot dispatch release: GitHub CLI 'gh' is not installed" >&2
  exit 1
fi

gh workflow run backend-release.yml \
  --repo "$repo" \
  -f "version=$version"

echo "Dispatched backend release workflow for backend/v$version."
echo "Check status with: gh run list --repo $repo --workflow backend-release.yml --limit 5"
