#!/bin/sh

set -eu

usage() {
  echo "usage: $0 [patch|minor|major]" >&2
}

bump=${1:-patch}

case "$bump" in
  patch|minor|major) ;;
  *)
    usage
    exit 2
    ;;
esac

latest_tag=$(git tag --list 'backend/v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)

if [ -z "$latest_tag" ]; then
  major=0
  minor=0
  patch=0
else
  version=${latest_tag#backend/v}
  base=${version%%-*}
  major=${base%%.*}
  rest=${base#*.}
  minor=${rest%%.*}
  patch=${rest#*.}
fi

case "$bump" in
  patch)
    patch=$((patch + 1))
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
esac

printf '%s.%s.%s\n' "$major" "$minor" "$patch"
