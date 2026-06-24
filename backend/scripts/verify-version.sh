#!/bin/sh

set -eu

if [ "$#" -ne 4 ]; then
  echo "usage: $0 <binary> <expected-version> <expected-commit> <expected-dirty>" >&2
  exit 2
fi

binary=$1
expected_version=${2#backend/v}
expected_version=${expected_version#v}
expected_commit=$3
expected_dirty=$4

if [ ! -x "$binary" ]; then
  echo "version verification failed: binary is not executable: $binary" >&2
  exit 1
fi

output=$($binary --version)

case "$output" in
  *"moh-sso-dashboard-backend $expected_version "*) ;;
  *)
    echo "version verification failed: expected version $expected_version" >&2
    echo "actual: $output" >&2
    exit 1
    ;;
esac

case "$output" in
  *"commit $expected_commit,"*) ;;
  *)
    echo "version verification failed: expected commit $expected_commit" >&2
    echo "actual: $output" >&2
    exit 1
    ;;
esac

case "$output" in
  *"dirty=$expected_dirty"*) ;;
  *)
    echo "version verification failed: expected dirty=$expected_dirty" >&2
    echo "actual: $output" >&2
    exit 1
    ;;
esac

case "$output" in
  *"vv$expected_version"*|*" dev "*|*"commit none,"*)
    echo "version verification failed: fallback or duplicate version metadata detected" >&2
    echo "actual: $output" >&2
    exit 1
    ;;
esac

echo "$output"
