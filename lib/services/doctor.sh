#!/usr/bin/env bash
#
# Android Security Research Lab
# Doctor Service (skeleton)
#

PASS_COUNT=0
WARN_COUNT=0
FAIL_COUNT=0

doctor_initialize() {
    PASS_COUNT=0
    WARN_COUNT=0
    FAIL_COUNT=0

    if [[ -n "${DOCTOR_CONFIG:-}" && -f "${DOCTOR_CONFIG}" ]]; then
        # shellcheck disable=SC1090
        source "${DOCTOR_CONFIG}"
    fi
}

doctor_record_pass() {
    doctor_pass "$1" "$2"
    PASS_COUNT=$((PASS_COUNT+1))
}

doctor_record_warn() {
    doctor_warn "$1" "$2"
    WARN_COUNT=$((WARN_COUNT+1))
}

doctor_record_fail() {
    doctor_fail "$1" "$2"
    FAIL_COUNT=$((FAIL_COUNT+1))
}

doctor_check_system() {
    info "System checks will be implemented in the next commit."
}

doctor_check_tools() {
    info "Tool checks will be implemented in the next commit."
}

doctor_execute() {
    doctor_check_system
    doctor_check_tools
}

doctor_summary() {
    section "Summary"
    summary_line "Passed" "${PASS_COUNT}"
    summary_line "Warnings" "${WARN_COUNT}"
    summary_line "Failed" "${FAIL_COUNT}"
}

doctor_service() {
    section "Android Framework Research Doctor"
    doctor_initialize
    doctor_execute
    doctor_summary
}
