#!/usr/bin/env bash

set -euo pipefail

TEST_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LAB="${TEST_ROOT}/bin/lab"

"${LAB}" --help | grep -q "Android Security Research Lab"
"${LAB}" --version | grep -q "Version :"
"${LAB}" bootstrap --help | grep -q "Usage: lab bootstrap"
"${LAB}" workspace --help | grep -q "Usage: lab workspace"

set +e
"${LAB}" unknown-command >/dev/null 2>&1
status=$?
set -e
[[ "${status}" -eq 2 ]]

set +e
"${LAB}" build >/dev/null 2>&1
status=$?
set -e
[[ "${status}" -eq 3 ]]

echo "cli_test: passed"
