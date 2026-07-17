#!/usr/bin/env bash

# ============================================================
# Local Lab Configuration
# ============================================================

LAB_CONFIG_LOADED=0

config_load() {
    if (( LAB_CONFIG_LOADED )); then
        return 0
    fi

    if [[ ! -e "${LAB_CONFIG_FILE}" ]]; then
        return 0
    fi

    if [[ ! -r "${LAB_CONFIG_FILE}" ]]; then
        error "Configuration file is not readable: ${LAB_CONFIG_FILE}"
        return 1
    fi

    # The local configuration is intentionally a trusted Bash file so paths
    # can reference LAB_ROOT and other environment variables.
    # shellcheck disable=SC1090
    source "${LAB_CONFIG_FILE}"
    LAB_CONFIG_LOADED=1
}

config_is_loaded() {
    (( LAB_CONFIG_LOADED ))
}

config_require() {
    local variable
    local missing=0

    for variable in "$@"; do
        if [[ ! "${variable}" =~ ^[A-Z_][A-Z0-9_]*$ ]]; then
            error "Invalid configuration key: ${variable}"
            return 2
        fi

        if [[ -z "${!variable:-}" ]]; then
            error "Missing required configuration: ${variable}"
            missing=1
        fi
    done

    return "${missing}"
}

config_prepare_output_dir() {
    mkdir -p "${OUTPUT_DIR}"
}
