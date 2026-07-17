#!/usr/bin/env bash

set -euo pipefail

BOOTSTRAP_MODE="plan"
BOOTSTRAP_INSTALLED_TOOLS=()
BOOTSTRAP_APT_PACKAGES=()
BOOTSTRAP_MANUAL_TOOLS=()
BOOTSTRAP_UNAVAILABLE_TOOLS=()

bootstrap_usage() {
    cat <<EOF
Usage: lab bootstrap [plan|--apply]

Modes:
    plan       Run the shared Doctor tool checks and show installation methods.
               This is the default and does not change the host.
    --apply    Install missing tools that have an apt installation method.
               Tools with only manual methods remain listed for the operator.
EOF
}

bootstrap_parse_args() {
    case "${1:-plan}" in
        plan) BOOTSTRAP_MODE="plan" ;;
        --apply) BOOTSTRAP_MODE="apply" ;;
        help|--help|-h) BOOTSTRAP_MODE="help" ;;
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

    [[ "${BOOTSTRAP_MODE}" == "help" ]] && return 0

    toolchain_load
    command_exists apt-get || { error "Bootstrap requires Ubuntu apt-get."; return 1; }
    command_exists apt-cache || { error "Bootstrap requires apt-cache."; return 1; }
}

bootstrap_add_apt_package() {
    local package="$1"
    local existing

    for existing in "${BOOTSTRAP_APT_PACKAGES[@]}"; do
        [[ "${existing}" == "${package}" ]] && return 0
    done

    BOOTSTRAP_APT_PACKAGES+=("${package}")
}

bootstrap_collect_install_plan() {
    local index package manual label status

    BOOTSTRAP_INSTALLED_TOOLS=()
    BOOTSTRAP_APT_PACKAGES=()
    BOOTSTRAP_MANUAL_TOOLS=()
    BOOTSTRAP_UNAVAILABLE_TOOLS=()

    toolchain_collect

    for (( index = 0; index < ${#TOOL_IDS[@]}; index++ )); do
        label="${TOOL_LABELS[index]}"
        status="${TOOL_STATUSES[index]}"

        if [[ "${status}" == "installed" ]]; then
            BOOTSTRAP_INSTALLED_TOOLS+=("${label}")
            continue
        fi

        if package=$(toolchain_apt_package_at "${index}") && apt-cache show "${package}" >/dev/null 2>&1; then
            bootstrap_add_apt_package "${package}"
            continue
        fi

        if manual=$(toolchain_manual_method_at "${index}"); then
            BOOTSTRAP_MANUAL_TOOLS+=("${label}: ${manual}")
        else
            BOOTSTRAP_UNAVAILABLE_TOOLS+=("${label}: no supported installation method")
        fi
    done
}

bootstrap_print_list() {
    local label="$1"
    shift

    (( $# == 0 )) && return 0
    info "${label}: $*"
}

bootstrap_print_plan() {
    section "Bootstrap Plan"

    info "Mode: ${BOOTSTRAP_MODE}"
    bootstrap_print_list "Already satisfied" "${BOOTSTRAP_INSTALLED_TOOLS[@]}"
    bootstrap_print_list "Install with apt" "${BOOTSTRAP_APT_PACKAGES[@]}"
    bootstrap_print_list "Manual action" "${BOOTSTRAP_MANUAL_TOOLS[@]}"
    bootstrap_print_list "Unavailable" "${BOOTSTRAP_UNAVAILABLE_TOOLS[@]}"

    if [[ "${BOOTSTRAP_MODE}" == "plan" ]]; then
        info "No changes were made. Run 'lab bootstrap --apply' to install available apt packages."
    fi
}

bootstrap_apply_packages() {
    if (( ${#BOOTSTRAP_APT_PACKAGES[@]} == 0 )); then
        success "No missing tools have an apt installation method."
        return 0
    fi

    info "Updating apt package metadata..."
    sudo apt-get update

    info "Installing missing tools with apt..."
    sudo apt-get install -y "${BOOTSTRAP_APT_PACKAGES[@]}"

    if (( ${#BOOTSTRAP_MANUAL_TOOLS[@]} > 0 )); then
        warn "Some tools still require manual installation. Review the Bootstrap Plan."
    fi
}

bootstrap_execute() {
    if [[ "${BOOTSTRAP_MODE}" == "help" ]]; then
        bootstrap_usage
        return 0
    fi

    bootstrap_collect_install_plan
    bootstrap_print_plan

    if [[ "${BOOTSTRAP_MODE}" == "apply" ]]; then
        bootstrap_apply_packages
    fi
}

bootstrap_summary() {
    [[ "${BOOTSTRAP_MODE}" == "help" ]] && return 0

    section "Bootstrap Summary"
    summary_line Satisfied "${#BOOTSTRAP_INSTALLED_TOOLS[@]}"
    summary_line AptPackages "${#BOOTSTRAP_APT_PACKAGES[@]}"
    summary_line Manual "${#BOOTSTRAP_MANUAL_TOOLS[@]}"
    summary_line Unavailable "${#BOOTSTRAP_UNAVAILABLE_TOOLS[@]}"
}

bootstrap_service() {
    bootstrap_initialize "$@"
    bootstrap_execute
    bootstrap_summary
}
