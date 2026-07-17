#!/usr/bin/env bash
set -euo pipefail

PASS_COUNT=0
WARN_COUNT=0
FAIL_COUNT=0

doctor_initialize() {
    PASS_COUNT=0
    WARN_COUNT=0
    FAIL_COUNT=0

    config_load
    [[ -f "${DOCTOR_CONFIG}" ]] && source "${DOCTOR_CONFIG}"
}

record_pass() {
    doctor_pass "$1" "$2"
    PASS_COUNT=$((PASS_COUNT + 1))
}

record_warn() {
    doctor_warn "$1" "$2"
    WARN_COUNT=$((WARN_COUNT + 1))
}

record_fail() {
    doctor_fail "$1" "$2"
    FAIL_COUNT=$((FAIL_COUNT + 1))
}

doctor_check_os() {
    source /etc/os-release

    case " ${SUPPORTED_UBUNTU} " in
        *" ${VERSION_ID} "*)
            record_pass "Ubuntu" "${VERSION_ID}"
            ;;
        *)
            record_fail "Ubuntu" "${VERSION_ID}"
            ;;
    esac
}

doctor_check_cpu() {
    local arch cores

    arch=$(uname -m)
    if [[ "${arch}" == "x86_64" ]]; then
        record_pass "Architecture" "${arch}"
    else
        record_warn "Architecture" "${arch}"
    fi

    cores=$(nproc)
    if (( cores >= RECOMMENDED_CPU_CORES )); then
        record_pass "CPU Cores" "${cores}"
    else
        record_warn "CPU Cores" "${cores} (recommend ${RECOMMENDED_CPU_CORES}+ )"
    fi
}

doctor_check_memory() {
    local mem

    mem=$(free -g | awk '/Mem:/{print $2}')
    if (( mem < MIN_MEMORY_GB )); then
        record_fail "Memory" "${mem} GB"
    elif (( mem < RECOMMENDED_MEMORY_GB )); then
        record_warn "Memory" "${mem} GB"
    else
        record_pass "Memory" "${mem} GB"
    fi
}

doctor_check_disk() {
    local free_gb

    free_gb=$(df -BG "${LAB_ROOT}" | awk 'NR==2 {gsub("G", "", $4); print $4}')
    if (( free_gb < MIN_DISK_GB )); then
        record_fail "Disk" "${free_gb} GB free"
    elif (( free_gb < RECOMMENDED_DISK_GB )); then
        record_warn "Disk" "${free_gb} GB free"
    else
        record_pass "Disk" "${free_gb} GB free"
    fi
}

doctor_check_lab_config() {
    if config_is_loaded; then
        record_pass "Lab Config" "${LAB_CONFIG_FILE}"
    else
        record_warn "Lab Config" "Not found (copy config/lab.conf.example)"
    fi
}

doctor_check_git() {
    if command -v git >/dev/null; then
        record_pass "Git" "$(git --version | awk '{print $3}')"
    else
        record_fail "Git" "Not Installed"
    fi
}

doctor_check_python() {
    if command -v python3 >/dev/null; then
        record_pass "Python3" "$(python3 --version | awk '{print $2}')"
    else
        record_fail "Python3" "Not Installed"
    fi
}

doctor_check_java() {
    if command -v java >/dev/null; then
        local v
        v=$(java -version 2>&1 | head -1 | sed -E 's/.*"?([0-9]+).*/\1/')
        if [ "${v}" -ge "${REQUIRED_JAVA_MAJOR}" ] 2>/dev/null; then
            record_pass "Java" "${v}"
        else
            record_warn "Java" "${v}"
        fi
    else
        record_fail "Java" "Not Installed"
    fi
}

doctor_check_repo() {
    if command -v repo >/dev/null; then
        record_pass "Repo" "Installed"
    else
        record_warn "Repo" "Not Installed"
    fi
}

doctor_check_ccache() {
    if command -v ccache >/dev/null; then
        record_pass "ccache" "Installed"
    else
        record_warn "ccache" "Recommended"
    fi
}

doctor_check_git_identity() {
    local n e

    n=$(git config --global user.name || true)
    e=$(git config --global user.email || true)
    if [[ -n "${n}" && -n "${e}" ]]; then
        record_pass "Git Identity" "${n} <${e}>"
    else
        record_warn "Git Identity" "Not Configured"
    fi
}

doctor_execute() {
    section "System Validation"

    doctor_check_os
    doctor_check_cpu
    doctor_check_memory
    doctor_check_disk
    doctor_check_lab_config

    doctor_check_git
    doctor_check_python
    doctor_check_java
    doctor_check_repo
    doctor_check_ccache
    doctor_check_git_identity
}

doctor_summary() {
    section "Summary"

    summary_line Passed "${PASS_COUNT}"
    summary_line Warnings "${WARN_COUNT}"
    summary_line Failed "${FAIL_COUNT}"
}

doctor_service() {
    section "Android Framework Research Doctor"

    doctor_initialize
    doctor_execute
    doctor_summary
}
