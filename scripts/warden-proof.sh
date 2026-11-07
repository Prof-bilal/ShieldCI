#!/usr/bin/env bash
set -euo pipefail

warden_bin="${1:?usage: warden-proof.sh /path/to/warden}"
policy="$(pwd)/backend/testdata/warden-fixture.yaml"
state_dir="$(mktemp -d)"
trap 'rm -rf "$state_dir"' EXIT
export XDG_STATE_HOME="$state_dir"

allowed_output="$($warden_bin run --policy "$policy" --backend linux -- /bin/echo shieldci-allowed)"
grep -Fxq 'shieldci-allowed' <<<"$allowed_output"

if $warden_bin run --policy "$policy" --backend linux -- /bin/cat /etc/passwd >/dev/null 2>&1; then
  echo 'Warden proof failed: undeclared /etc/passwd access was allowed' >&2
  exit 1
fi

echo 'Warden proof passed: allowed command succeeded and undeclared filesystem access was blocked.'
