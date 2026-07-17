#!/usr/bin/env bash

set -euo pipefail

BOOTSTRAP_MODE="plan"
BOOTSTRAP_INSTALLED_PACKAGES=()
BOOTSTRAP_MISSING_PACKAGES=()
BOOTSTRAP_UNAVAILABLE_PACKAGES=()

bootstrap_usage() {
    cat <<EOF
Usage: lab bootstrap [plan|--apply]

Modes:
    plan       Show packages that are installed, missing, or unavailable.
               This is the default and does not change the host.
    --apply    Update apt metadata and install all missing packages.
EOF
}

bootstrap_parse_args() {
    case "${1:-plan}" in
        plan)
            BOOTSTRAP_MODE="plan"
            ;;
        --apply)
            BOOTSTRAP_MODE="apply"
            ;;
        help|--help|-h)
            BOOTSTRAP_MODE="help"
            ;;
        *)
            error "Unknown bootstrap option: ${1}"
            bootstrap_usage >&2
            return 2
            ;;
    esac

    if (( $# > 1 )); then
        error "Bootstrap accepts at most one option."
        bootstrap_usage >&2
        return 2
    fi
}

bootstrap_initialize() {
    bootstrap_parse_args "$@"

    if [[ "${BOOTSTRAP_MODE}" == "help" ]]; then
        return 0
    fi

    if [[ ! -r "${BOOTSTRAP_CONFIG}" ]]; then
        error "Bootstrap configuration is not readable: ${BOOTSTRAP_CONFIG}"
        return 1
    fi

    # shellcheck disable=SC1090
    source "${BOOTSTRAP_CONFIG}"

    if ! declare -p BOOTSTRAP_APT_PACKAGES >/dev/null 2>&1; then
        error "BOOTSTRAP_APT_PACKAGES is not defined in ${BOOTSTRAP_CONFIG}"
        return 1
    fi

    command_exists apt-get || { error "Bootstrap requires Ubuntu apt-get."; return 1; }
    command_exists apt-cache || { error "Bootstrap requires apt-cache."; return 1; }
    command_exists dpkg-query || { error "Bootstrap requires dpkg-query."; return 1; }
}

bootstrap_package_is_installed() {
    dpkg-query -W -f='${db:Status-Status}' "$1" 2>/dev/null | grep -qx "installed"
}

bootstrap_package_is_available() {
    apt-cache show "$1" >/dev/null 2>&1
}

bootstrap_collect_package_state() {
    local package

    BOOTSTRAP_INSTALLED_PACKAGES=()
    BOOTSTRAP_MISSING_PACKAGES=()
    BOOTSTRAP_UNAVAILABLE_PACKAGES=()

    for package in "${BOOTSTRAP_APT_PACKAGES[@]}"; do
        if bootstrap_package_is_installed "${package}"; then
            BOOTSTRAP_INSTALLED_PACKAGES+=("${package}")
        elif bootstrap_package_is_available "${package}"; then
            BOOTSTRAP_MISSING_PACKAGES+=("${package}")
        else
            BOOTSTRAP_UNAVAILABLE_PACKAGES+=("${package}")
        fi
    done
}

bootstrap_print_package_list() {
    local label="$1"
    shift

    (( $# == 0 )) && return 0
    info "${label}: $*"
}

bootstrap_print_plan() {
    section "Bootstrap Plan"

    info "Mode: ${BOOTSTRAP_MODE}"
    bootstrap_print_package_list "Already installed" "${BOOTSTRAP_INSTALLED_PACKAGES[@]}"
    bootstrap_print_package_list "Will install" "${BOOTSTRAP_MISSING_PACKAGES[@]}"
    bootstrap_print_package_list "Unavailable" "${BOOTSTRAP_UNAVAILABLE_PACKAGES[@]}"

    if [[ "${BOOTSTRAP_MODE}" == "plan" ]]; then
        info "No changes were made. Run 'lab bootstrap --apply' to install missing packages."
    fi
}

bootstrap_apply_packages() {
    if (( ${#BOOTSTRAP_UNAVAILABLE_PACKAGES[@]} > 0 )); then
        error "Cannot install while required packages are unavailable."
        return 1
    fi

    if (( ${#BOOTSTRAP_MISSING_PACKAGES[@]} == 0 )); then
        success "All bootstrap packages are already installed."
        return 0
    fi

    info "Updating apt package metadata..."
    sudo apt-get update

    info "Installing missing bootstrap packages..."
    sudo apt-get install -y "${BOOTSTRAP_MISSING_PACKAGES[@]}"
}

bootstrap_execute() {
    if [[ "${BOOTSTRAP_MODE}" == "help" ]]; then
        bootstrap_usage
        return 0
    fi

    bootstrap_collect_package_state
    bootstrap_print_plan

    if [[ "${BOOTSTRAP_MODE}" == "apply" ]]; then
        bootstrap_apply_packages
    fi
}

bootstrap_summary() {
    [[ "${BOOTSTRAP_MODE}" == "help" ]] && return 0

    section "Bootstrap Summary"
    summary_line Installed "${#BOOTSTRAP_INSTALLED_PACKAGES[@]}"
    summary_line Missing "${#BOOTSTRAP_MISSING_PACKAGES[@]}"
    summary_line Unavailable "${#BOOTSTRAP_UNAVAILABLE_PACKAGES[@]}"
}

bootstrap_service() {
    bootstrap_initialize "$@"
    bootstrap_execute
    bootstrap_summary
}
