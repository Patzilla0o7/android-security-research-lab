#!/usr/bin/env bash

set -euo pipefail

TEST_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export LAB_ROOT="${TEST_ROOT}"

source "${LAB_ROOT}/lib/core/constants.sh"
source "${LAB_ROOT}/lib/core/logger.sh"
source "${LAB_ROOT}/lib/core/utils.sh"
source "${LAB_ROOT}/lib/core/toolchain.sh"

toolchain_load

[[ "$(toolchain_tool_count)" -ge 22 ]]

java_index=$(toolchain_find_index java)
[[ "${TOOL_CHECK_TYPES[java_index]}" == "java-major" ]]
[[ "${TOOL_MIN_VERSIONS[java_index]}" == "17" ]]
[[ "$(toolchain_apt_package_at "${java_index}")" == "openjdk-17-jdk" ]]
[[ "$(printf '%s\n' 'openjdk version "17.0.19" 2026-04-21' | toolchain_java_major)" == "17" ]]

repo_index=$(toolchain_find_index repo)
[[ "$(toolchain_apt_package_at "${repo_index}")" == "repo" ]]
[[ "$(toolchain_manual_method_at "${repo_index}")" == "https://gerrit.googlesource.com/git-repo" ]]

adb_index=$(toolchain_find_index adb)
[[ "$(toolchain_apt_package_at "${adb_index}")" == "adb" ]]

emulator_index=$(toolchain_find_index emulator)
[[ "$(toolchain_manual_method_at "${emulator_index}")" == "https://developer.android.com/studio/run/emulator-commandline" ]]

echo "toolchain_test: passed"
