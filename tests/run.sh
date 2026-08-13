#!/usr/bin/env bash

set -euo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(cd "${TEST_DIR}/.." && pwd)"
if ! command -v go >/dev/null 2>&1; then
    echo "Go 1.22 or newer is required to run tests." >&2
    exit 1
fi

(
    cd "${TEST_ROOT}"
    go test ./...
)

"${TEST_DIR}/cli_test.sh"
