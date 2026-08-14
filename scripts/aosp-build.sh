#!/usr/bin/env bash

# AOSP envsetup.sh intentionally probes variables such as TOP before defining
# them, so nounset is incompatible with the environment it provides.
set -eo pipefail

if [[ "$#" -lt 4 ]]; then
    echo "Usage: aosp-build.sh <workspace> <target> <jobs> <ccache-dir> [module ...]" >&2
    exit 2
fi

workspace="$1"
target="$2"
jobs="$3"
ccache_dir="$4"
shift 4
cd "$workspace"

if [[ -n "$ccache_dir" ]]; then
    mkdir -p "$ccache_dir"
    export USE_CCACHE=1
    export CCACHE_DIR="$ccache_dir"
fi

# AOSP defines lunch and m as shell functions. This is the intentionally
# retained Shell boundary of the Go-controlled build workflow.
source build/envsetup.sh
lunch "$target"

if [[ "$#" -gt 0 ]]; then
    m -j"$jobs" "$@"
else
    m -j"$jobs"
fi

if command -v ccache >/dev/null 2>&1; then
    echo
    echo "CCache statistics"
    ccache -s || true
fi
