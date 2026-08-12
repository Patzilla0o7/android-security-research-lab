#!/usr/bin/env bash

set -euo pipefail

LAB_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v go >/dev/null 2>&1; then
    echo "Go 1.22 or newer is required to build ASRL." >&2
    exit 1
fi

mkdir -p "${LAB_ROOT}/build"
go build -o "${LAB_ROOT}/build/lab" "${LAB_ROOT}/cmd/lab"
echo "Built ${LAB_ROOT}/build/lab"
