#!/bin/sh

set -eu

usage() {
  echo "usage: $0 <version> [repo] [image]" >&2
  echo "example: $0 1.2.3 mohuganda/moh-sso-dashboard ghcr.io/mohuganda/moh-sso-dashboard-backend" >&2
}

if [ "$#" -lt 1 ] || [ "$#" -gt 3 ]; then
  usage
  exit 2
fi

version=${1#backend/v}
version=${version#v}
repo=${2:-mohuganda/moh-sso-dashboard}
image=${3:-ghcr.io/mohuganda/moh-sso-dashboard-backend}
tag="backend/v$version"

if ! command -v gh >/dev/null 2>&1; then
  echo "cannot inspect GitHub release: GitHub CLI 'gh' is not installed" >&2
  exit 1
fi

echo "GitHub release:"
gh release view "$tag" --repo "$repo"

echo
echo "Recent backend release workflow runs:"
gh run list --repo "$repo" --workflow backend-release.yml --limit 5

echo
echo "Recent backend release PR workflow runs:"
gh run list --repo "$repo" --workflow backend-release-pr.yml --limit 5

if command -v docker >/dev/null 2>&1; then
  echo
  echo "Docker image manifest:"
  docker buildx imagetools inspect "$image:$version"
else
  echo
  echo "Docker CLI not found; skipping image manifest inspection."
fi
