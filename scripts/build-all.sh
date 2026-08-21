#!/usr/bin/env bash

set -euo pipefail

ASRL_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DROIDFORGE_ROOT="${ASRL_ROOT}/tools/DroidForge"

if ! command -v go >/dev/null 2>&1; then
    echo "Go 1.22 or newer is required to build ASRL and DroidForge." >&2
    exit 1
fi

GO_VERSION="$(go env GOVERSION)"
GO_VERSION="${GO_VERSION#go}"
GO_MAJOR="${GO_VERSION%%.*}"
GO_REMAINDER="${GO_VERSION#*.}"
GO_MINOR="${GO_REMAINDER%%.*}"
if ((GO_MAJOR < 1 || (GO_MAJOR == 1 && GO_MINOR < 22))); then
    echo "Go 1.22 or newer is required to build ASRL and DroidForge (found $(go env GOVERSION))." >&2
    exit 1
fi

if ! command -v git >/dev/null 2>&1; then
    echo "Git is required to initialize DroidForge." >&2
    exit 1
fi

git -C "${ASRL_ROOT}" submodule update --init --recursive -- tools/DroidForge
if [[ ! -f "${DROIDFORGE_ROOT}/Makefile" ]]; then
    echo "DroidForge submodule is not initialized: ${DROIDFORGE_ROOT}" >&2
    exit 1
fi

"${ASRL_ROOT}/scripts/build-go.sh"
make -C "${DROIDFORGE_ROOT}" build

echo "Built ${ASRL_ROOT}/bin/lab"
echo "Built ${DROIDFORGE_ROOT}/bin/droidforge"
