#!/usr/bin/env bash

set -euo pipefail

TEST_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export LAB_ROOT="${TEST_ROOT}"
export LAB_CONFIG_FILE="${TEST_ROOT}/config/lab.conf.example"

source "${LAB_ROOT}/lib/core/constants.sh"
source "${LAB_ROOT}/lib/core/logger.sh"
source "${LAB_ROOT}/lib/core/config.sh"

config_load

[[ "${ANDROID_WORKSPACE}" == "${LAB_ROOT}/workspace/aosp" ]]
[[ "${ANDROID_BRANCH}" == "android-latest-release" ]]
[[ "${JAVA_HOME}" == "/usr/lib/jvm/java-17-openjdk-amd64" ]]
config_require ANDROID_WORKSPACE ANDROID_MANIFEST_URL ANDROID_BRANCH ANDROID_BUILD_TARGET JAVA_HOME

if config_require MISSING_CONFIG_VALUE >/dev/null 2>&1; then
    echo "config_require accepted a missing key" >&2
    exit 1
fi

echo "config_test: passed"
