#!/usr/bin/env bash
# Runs govulncheck and fails on any *reachable* advisory that is not listed in
# .govulncheck-ignore.
#
# Standard-library advisories are a property of the toolchain that compiles
# the code, which for a library is the consumer's toolchain, not go.mod's
# minimum. CI therefore runs this gate twice:
#   - on the declared minimum toolchain with GOVULNCHECK_SKIP_STDLIB=1, which
#     enforces module (dependency) advisories and only reports stdlib ones;
#   - on the current stable toolchain with full enforcement.
set -euo pipefail

ignore_file="${1:-.govulncheck-ignore}"
skip_stdlib="${GOVULNCHECK_SKIP_STDLIB:-0}"
out="$(mktemp)"
trap 'rm -f "$out"' EXIT

# govulncheck exits 0 (clean) or 3 (vulnerabilities found). Any other status
# is an operational failure (missing binary, package load error, crash) and
# must fail the gate rather than be mistaken for a clean scan.
set +e
govulncheck -format json ./... > "$out"
status=$?
set -e
if [ "$status" -ne 0 ] && [ "$status" -ne 3 ]; then
  echo "::error::govulncheck failed with exit status $status"
  cat "$out"
  exit "$status"
fi
if ! [ -s "$out" ] || ! jq -e 'select(.config != null)' "$out" > /dev/null 2>&1; then
  echo "::error::govulncheck produced no usable JSON output"
  cat "$out"
  exit 1
fi

# "finding" objects whose first trace frame names a function are the ones
# where the vulnerable symbol is actually called (reachable).
reachable_all="$(jq -r 'select(.finding != null) | .finding | select(.trace[0].function != null) | .osv' "$out" | sort -u)"
# OSV entries whose affected module is the standard library.
stdlib_ids="$(jq -r 'select(.osv != null) | .osv | select([.affected[].package.name] | index("stdlib")) | .id' "$out" | sort -u)"
if [ "$skip_stdlib" = "1" ]; then
  reachable="$(comm -23 <(echo "$reachable_all") <(echo "$stdlib_ids") | sed '/^$/d' || true)"
  skipped="$(comm -12 <(echo "$reachable_all") <(echo "$stdlib_ids") | sed '/^$/d' || true)"
  if [ -n "$skipped" ]; then
    echo "::warning::reachable standard-library advisories (not enforced on the minimum toolchain; fixed by a newer Go release):"
    echo "$skipped"
  fi
else
  reachable="$reachable_all"
fi
accepted="$(grep -vE '^\s*(#|$)' "$ignore_file" | awk '{print $1}' | sort -u || true)"

unexpected="$(comm -23 <(echo "$reachable") <(echo "$accepted") | sed '/^$/d' || true)"
if [ -n "$unexpected" ]; then
  echo "::error::govulncheck found reachable vulnerabilities not in $ignore_file:"
  echo "$unexpected"
  echo
  govulncheck ./... || true
  exit 1
fi

echo "govulncheck: no unaccepted reachable vulnerabilities."
if [ -n "$reachable" ]; then
  echo "Accepted (listed in $ignore_file):"
  echo "$reachable"
fi
