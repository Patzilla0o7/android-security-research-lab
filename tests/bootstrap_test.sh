#!/usr/bin/env bash

set -euo pipefail

TEST_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export LAB_ROOT="${TEST_ROOT}"

source "${LAB_ROOT}/lib/core/constants.sh"
source "${LAB_ROOT}/lib/core/logger.sh"
source "${LAB_ROOT}/lib/core/utils.sh"
source "${LAB_ROOT}/lib/core/toolchain.sh"
source "${LAB_ROOT}/lib/services/bootstrap.sh"

bootstrap_parse_args plan
[[ "${BOOTSTRAP_MODE}" == "plan" ]]

bootstrap_parse_args --apply
[[ "${BOOTSTRAP_MODE}" == "apply" ]]

bootstrap_parse_args --help
[[ "${BOOTSTRAP_MODE}" == "help" ]]

if bootstrap_parse_args invalid >/dev/null 2>&1; then
    echo "bootstrap_parse_args accepted an invalid option" >&2
    exit 1
fi

echo "bootstrap_test: passed"
