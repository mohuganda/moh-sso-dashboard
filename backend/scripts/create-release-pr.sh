#!/bin/sh

set -eu

usage() {
  echo "usage: $0 <version> [repo] [base] [remote]" >&2
  echo "example: $0 1.2.3 mohuganda/moh-sso-dashboard main upstream" >&2
}

remote_owner() {
  remote_name=$1
  remote_url=$(git remote get-url "$remote_name" 2>/dev/null || true)

  case "$remote_url" in
    git@github.com:*.git)
      owner_repo=${remote_url#git@github.com:}
      owner_repo=${owner_repo%.git}
      printf '%s\n' "${owner_repo%%/*}"
      ;;
    https://github.com/*/*.git)
      owner_repo=${remote_url#https://github.com/}
      owner_repo=${owner_repo%.git}
      printf '%s\n' "${owner_repo%%/*}"
      ;;
    https://github.com/*/*)
      owner_repo=${remote_url#https://github.com/}
      printf '%s\n' "${owner_repo%%/*}"
      ;;
    *)
      printf '\n'
      ;;
  esac
}

if [ "$#" -lt 1 ] || [ "$#" -gt 4 ]; then
  usage
  exit 2
fi

version=${1#backend/v}
version=${version#v}
repo=${2:-mohuganda/moh-sso-dashboard}
base=${3:-main}
remote=${4:-upstream}

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

target_remote=${repo%%/*}
base_ref="$base"
if git show-ref --verify --quiet "refs/remotes/$target_remote/$base"; then
  base_ref="$target_remote/$base"
elif git show-ref --verify --quiet "refs/remotes/upstream/$base"; then
  base_ref="upstream/$base"
elif git show-ref --verify --quiet "refs/remotes/origin/$base"; then
  base_ref="origin/$base"
fi

if git rev-parse --verify "$base_ref" >/dev/null 2>&1; then
  commits_ahead=$(git rev-list --count "$base_ref..HEAD")
  if [ "$commits_ahead" = "0" ]; then
    git commit --allow-empty -m "Prepare backend v$version release"
  fi
else
  echo "warning: could not resolve base ref '$base_ref'; skipping empty release marker check" >&2
fi

git push -u "$remote" "$branch"

head_owner=$(remote_owner "$remote")
head_ref="$branch"
if [ -n "$head_owner" ] && [ "$head_owner" != "${repo%%/*}" ]; then
  echo "warning: release PR is being opened from fork owner '$head_owner' into '$repo'." >&2
  echo "warning: GitHub will not allow the fork PR workflow token to create release tags in '$repo'." >&2
  echo "warning: after merging this PR, run: make release-dispatch VERSION=$version REPO=$repo" >&2
  head_ref="$head_owner:$branch"
fi

gh pr create \
  --repo "$repo" \
  --base "$base" \
  --head "$head_ref" \
  --title "Release backend v$version" \
  --body "Prepares backend release $tag."
