#!/bin/sh

set -eu

usage() {
  echo "usage: $0 <version> [repo] [base] [remote]" >&2
  echo "example: $0 1.2.3 mohuganda/moh-sso-dashboard main origin" >&2
}

if [ "$#" -lt 1 ] || [ "$#" -gt 4 ]; then
  usage
  exit 2
fi

version=${1#backend/v}
version=${version#v}
repo=${2:-mohuganda/moh-sso-dashboard}
base=${3:-main}
remote=${4:-origin}

if ! printf '%s' "$version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'; then
  echo "invalid backend semantic version: $version" >&2
  exit 1
fi

branch="release/backend-v$version"
tag="backend/v$version"

if git rev-parse "refs/tags/$tag" >/dev/null 2>&1; then
  echo "cannot create release PR: tag already exists: $tag" >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "cannot create release PR: GitHub CLI 'gh' is not installed" >&2
  exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
  echo "cannot create release PR: commit or stash local changes first" >&2
  git status --short
  exit 1
fi

current_branch=$(git branch --show-current)

if [ "$current_branch" != "$branch" ]; then
  if git show-ref --verify --quiet "refs/heads/$branch"; then
    git checkout "$branch"
  else
    git checkout -b "$branch"
  fi
fi

git push -u "$remote" "$branch"

gh pr create \
  --repo "$repo" \
  --base "$base" \
  --head "$branch" \
  --title "Release backend v$version" \
  --body "Prepares backend release $tag."
