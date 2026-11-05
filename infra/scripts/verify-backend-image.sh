#!/bin/sh

set -eu

usage() {
  echo "usage: $0 <image> <expected-version> [expected-commit]" >&2
  echo "example: $0 ghcr.io/mohuganda/moh-sso-dashboard-backend:1.2.3 1.2.3 abc123" >&2
}

if [ "$#" -lt 2 ] || [ "$#" -gt 3 ]; then
  usage
  exit 2
fi

image=$1
expected_version=${2#backend/v}
expected_version=${expected_version#v}
expected_commit=${3:-}

if ! command -v docker >/dev/null 2>&1; then
  echo "cannot verify image: Docker CLI is not installed" >&2
  exit 1
fi

output=$(docker run --rm "$image" --version)
echo "$output"

case "$output" in
  *"moh-sso-dashboard-backend $expected_version "*) ;;
  *)
    echo "image verification failed: expected version $expected_version" >&2
    exit 1
    ;;
esac

if [ -n "$expected_commit" ]; then
  case "$output" in
    *"commit $expected_commit,"*) ;;
    *)
      echo "image verification failed: expected commit $expected_commit" >&2
      exit 1
      ;;
  esac
fi

case "$output" in
  *" dev "*|*"commit none,"*|*"dirty=true"*)
    echo "image verification failed: release image contains fallback or dirty metadata" >&2
    exit 1
    ;;
esac

# Exercise the final image as the same unprivileged user used by the entrypoint.
# --entrypoint avoids needing production configuration or mounted secrets.
docker run --rm --user appuser --entrypoint sh "$image" -ec '
  test "$PWD" = /app
  for filename in uganda_districts.geojson uganda_subcounties.geojson uga_water.geojson; do
    path="./assets/geojson/$filename"
    if [ ! -r "$path" ] || [ ! -s "$path" ]; then
      echo "image verification failed: missing, empty or unreadable runtime asset: $path" >&2
      exit 1
    fi
  done
'

echo "Backend image metadata and runtime GeoJSON assets verified."
