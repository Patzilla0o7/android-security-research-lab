#!/usr/bin/env bash

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

require_command() {
    if ! command_exists "$1"; then
        error "Required command is not available: $1"
        return "${EXIT_FAILURE}"
    fi
}

require_absolute_path() {
    local label="$1"
    local path="$2"

    if [[ -z "${path}" || "${path}" != /* ]]; then
        error "${label} must be an absolute path: ${path:-<empty>}"
        return "${EXIT_USAGE}"
    fi
}

path_is_within() {
    local path="${1%/}"
    local parent="${2%/}"

    [[ "${path}" == "${parent}" || "${path}" == "${parent}/"* ]]
}

refuse_dangerous_path() {
    local label="$1"
    local path="${2%/}"

    require_absolute_path "${label}" "${path}" || return

    case "${path}" in
        ""|/|"${LAB_ROOT}"|"${LAB_ROOT}/config"|"${LAB_ROOT}/docs"|"${LAB_ROOT}/research")
            error "Refusing unsafe ${label}: ${path}"
            return "${EXIT_FAILURE}"
            ;;
    esac
}
