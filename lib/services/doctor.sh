#!/usr/bin/env bash

#
# ============================================================
#
# Android Security Research Lab
#
# Doctor Service
#
# ============================================================
#

PASS_COUNT=0
WARN_COUNT=0
FAIL_COUNT=0

doctor_initialize() {

    PASS_COUNT=0
    WARN_COUNT=0
    FAIL_COUNT=0

    if [[ -f "${DOCTOR_CONFIG}" ]]; then
        # shellcheck disable=SC1090
        source "${DOCTOR_CONFIG}"
    fi

}

doctor_pass() {

    success "$1"

    PASS_COUNT=$((PASS_COUNT + 1))

}

doctor_warn() {

    warn "$1"

    WARN_COUNT=$((WARN_COUNT + 1))

}

doctor_fail() {

    error "$1"

    FAIL_COUNT=$((FAIL_COUNT + 1))

}

doctor_execute() {

    info "Doctor checks are not implemented yet."

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