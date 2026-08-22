#!/usr/bin/env bash
# Runs govulncheck and fails on any *reachable* advisory that is not listed in
# .govulncheck-ignore. Standard-library advisories are compared against the
# toolchain in use, so keeping CI on the latest patch release keeps them clear.
set -euo pipefail

ignore_file="${1:-.govulncheck-ignore}"
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
reachable="$(jq -r 'select(.finding != null) | .finding | select(.trace[0].function != null) | .osv' "$out" | sort -u)"
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
