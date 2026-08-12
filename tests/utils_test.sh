#!/usr/bin/env bash

set -euo pipefail

TEST_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export LAB_ROOT="${TEST_ROOT}"

source "${LAB_ROOT}/lib/core/constants.sh"
source "${LAB_ROOT}/lib/core/logger.sh"
source "${LAB_ROOT}/lib/core/utils.sh"

require_absolute_path "test path" "/tmp/asrl-test"
path_is_within "/tmp/asrl-test/child" "/tmp/asrl-test"
refuse_dangerous_path "test path" "/tmp/asrl-test"

if require_absolute_path "test path" "relative/path" >/dev/null 2>&1; then
    echo "require_absolute_path accepted a relative path" >&2
    exit 1
fi

if refuse_dangerous_path "test path" "${LAB_ROOT}" >/dev/null 2>&1; then
    echo "refuse_dangerous_path accepted LAB_ROOT" >&2
    exit 1
fi

echo "utils_test: passed"
