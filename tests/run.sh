#!/usr/bin/env bash

set -euo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"${TEST_DIR}/config_test.sh"
"${TEST_DIR}/bootstrap_test.sh"
