#!/usr/bin/env bash

# ============================================================
# Shared Toolchain Registry and Detection
# ============================================================

TOOL_IDS=()
TOOL_LABELS=()
TOOL_LEVELS=()
TOOL_CHECK_TYPES=()
TOOL_CHECK_TARGETS=()
TOOL_MIN_VERSIONS=()
TOOL_INSTALL_METHODS=()
TOOL_STATUSES=()
TOOL_DETAILS=()
TOOLCHAIN_LOADED=0

toolchain_load() {
    local spec id label level check_type check_target min_version methods

    if (( TOOLCHAIN_LOADED )); then
        return 0
    fi

    if [[ ! -r "${TOOLCHAIN_CONFIG}" ]]; then
        error "Toolchain configuration is not readable: ${TOOLCHAIN_CONFIG}"
        return 1
    fi

    # shellcheck disable=SC1090
    source "${TOOLCHAIN_CONFIG}"

    if ! declare -p TOOL_SPECS >/dev/null 2>&1; then
        error "TOOL_SPECS is not defined in ${TOOLCHAIN_CONFIG}"
        return 1
    fi

    for spec in "${TOOL_SPECS[@]}"; do
        IFS='|' read -r id label level check_type check_target min_version methods <<< "${spec}"

        if [[ -z "${id}" || -z "${label}" || -z "${level}" || -z "${check_type}" || -z "${check_target}" || -z "${methods}" ]]; then
            error "Invalid tool specification: ${spec}"
            return 1
        fi

        case "${level}" in
            required|recommended|optional) ;;
            *) error "Invalid tool level for ${id}: ${level}"; return 1 ;;
        esac

        case "${check_type}" in
            command|java-major|package) ;;
            *) error "Invalid check type for ${id}: ${check_type}"; return 1 ;;
        esac

        TOOL_IDS+=("${id}")
        TOOL_LABELS+=("${label}")
        TOOL_LEVELS+=("${level}")
        TOOL_CHECK_TYPES+=("${check_type}")
        TOOL_CHECK_TARGETS+=("${check_target}")
        TOOL_MIN_VERSIONS+=("${min_version}")
        TOOL_INSTALL_METHODS+=("${methods}")
    done

    TOOLCHAIN_LOADED=1
}

toolchain_tool_count() {
    echo "${#TOOL_IDS[@]}"
}

toolchain_find_index() {
    local id="$1"
    local index

    for (( index = 0; index < ${#TOOL_IDS[@]}; index++ )); do
        if [[ "${TOOL_IDS[index]}" == "${id}" ]]; then
            echo "${index}"
            return 0
        fi
    done

    return 1
}

toolchain_command_detail() {
    local id="$1"
    local command="$2"

    case "${id}" in
        git) git --version | awk '{print $3}' ;;
        python3) python3 --version | awk '{print $2}' ;;
        ccache) ccache --version | awk 'NR==1 {print $3}' ;;
        repo) repo version 2>/dev/null | awk -F': ' '/repo launcher version/ {print $2; exit}' ;;
        *) echo "Installed" ;;
    esac
}

toolchain_java_major() {
    sed -nE 's/.*version "([0-9]+).*/\1/p'
}

toolchain_check_at() {
    local index="$1"
    local id="${TOOL_IDS[index]}"
    local check_type="${TOOL_CHECK_TYPES[index]}"
    local check_target="${TOOL_CHECK_TARGETS[index]}"
    local min_version="${TOOL_MIN_VERSIONS[index]}"
    local version

    case "${check_type}" in
        command)
            if command_exists "${check_target}"; then
                TOOL_STATUSES[index]="installed"
                TOOL_DETAILS[index]="$(toolchain_command_detail "${id}" "${check_target}")"
            else
                TOOL_STATUSES[index]="missing"
                TOOL_DETAILS[index]="Not installed"
            fi
            ;;
        java-major)
            if ! command_exists "${check_target}"; then
                TOOL_STATUSES[index]="missing"
                TOOL_DETAILS[index]="Not installed"
                return 0
            fi

            version=$(java -version 2>&1 | head -1 | toolchain_java_major)
            if [[ ! "${version}" =~ ^[0-9]+$ ]]; then
                TOOL_STATUSES[index]="version_mismatch"
                TOOL_DETAILS[index]="Unable to determine version"
            elif (( version < min_version )); then
                TOOL_STATUSES[index]="version_mismatch"
                TOOL_DETAILS[index]="${version} (requires ${min_version}+)"
            else
                TOOL_STATUSES[index]="installed"
                TOOL_DETAILS[index]="${version}"
            fi
            ;;
        package)
            if dpkg-query -W -f='${db:Status-Status}' "${check_target}" 2>/dev/null | grep -qx "installed"; then
                TOOL_STATUSES[index]="installed"
                TOOL_DETAILS[index]="Installed"
            else
                TOOL_STATUSES[index]="missing"
                TOOL_DETAILS[index]="Not installed"
            fi
            ;;
    esac
}

toolchain_collect() {
    local index

    toolchain_load
    TOOL_STATUSES=()
    TOOL_DETAILS=()

    for (( index = 0; index < ${#TOOL_IDS[@]}; index++ )); do
        toolchain_check_at "${index}"
    done
}

toolchain_apt_package_at() {
    local index="$1"
    local method
    local methods="${TOOL_INSTALL_METHODS[index]}"

    IFS=',' read -r -a method_list <<< "${methods}"
    for method in "${method_list[@]}"; do
        if [[ "${method}" == apt:* ]]; then
            echo "${method#apt:}"
            return 0
        fi
    done

    return 1
}

toolchain_manual_method_at() {
    local index="$1"
    local method
    local methods="${TOOL_INSTALL_METHODS[index]}"

    IFS=',' read -r -a method_list <<< "${methods}"
    for method in "${method_list[@]}"; do
        if [[ "${method}" == manual:* ]]; then
            echo "${method#manual:}"
            return 0
        fi
    done

    return 1
}
