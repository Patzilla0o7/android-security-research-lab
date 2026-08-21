#!/usr/bin/env bash

set -euo pipefail

TEST_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LAB="${TEST_ROOT}/bin/lab"

"${LAB}" --help | grep -q "Android Security Research Lab"
"${LAB}" --version | grep -q "Version :"
"${LAB}" bootstrap --help | grep -q "Usage: lab bootstrap"
"${LAB}" workspace --help | grep -q "Usage: lab workspace"
"${LAB}" repo --help | grep -q "Usage: lab repo"
"$LAB" build --help | grep -q "Usage: lab build"
"$LAB" device --help | grep -q "Usage: lab device"
"$LAB" collect --help | grep -q "Usage: lab collect"
"$LAB" research --help | grep -q "Usage: lab research"

set +e
"${LAB}" unknown-command >/dev/null 2>&1
status=$?
set -e
[[ "${status}" -eq 2 ]]

echo "cli_test: passed"
