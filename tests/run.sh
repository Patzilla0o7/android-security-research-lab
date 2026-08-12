#!/usr/bin/env bash

set -euo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(cd "${TEST_DIR}/.." && pwd)"

if command -v go >/dev/null 2>&1; then
    (
        cd "${TEST_ROOT}"
        go test ./...
    )
else
    echo "go_test: skipped (Go toolchain is not installed)"
fi


"${TEST_DIR}/config_test.sh"
"${TEST_DIR}/toolchain_test.sh"
"${TEST_DIR}/bootstrap_test.sh"
"${TEST_DIR}/utils_test.sh"
"${TEST_DIR}/cli_test.sh"
