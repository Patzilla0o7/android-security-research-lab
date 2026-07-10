#!/usr/bin/env bash

PASS_COUNT=0
WARN_COUNT=0
FAIL_COUNT=0

doctor_initialize() {
    PASS_COUNT=0
    WARN_COUNT=0
    FAIL_COUNT=0
    [[ -f "${DOCTOR_CONFIG}" ]] && source "${DOCTOR_CONFIG}"
}

record_pass(){ doctor_pass "$1" "$2"; PASS_COUNT=$((PASS_COUNT+1)); }
record_warn(){ doctor_warn "$1" "$2"; WARN_COUNT=$((WARN_COUNT+1)); }
record_fail(){ doctor_fail "$1" "$2"; FAIL_COUNT=$((FAIL_COUNT+1)); }

doctor_check_os(){
    source /etc/os-release
    case " ${SUPPORTED_UBUNTU} " in
      *" ${VERSION_ID} "*) record_pass "Ubuntu" "${VERSION_ID}" ;;
      *) record_fail "Ubuntu" "${VERSION_ID}" ;;
    esac
}

doctor_check_cpu(){
    arch=$(uname -m)
    [[ "$arch" == "x86_64" ]] && record_pass "Architecture" "$arch" || record_warn "Architecture" "$arch"
    cores=$(nproc)
    if (( cores >= RECOMMENDED_CPU_CORES )); then
      record_pass "CPU Cores" "$cores"
    else
      record_warn "CPU Cores" "$cores (recommend ${RECOMMENDED_CPU_CORES}+)"
    fi
}

doctor_check_memory(){
    mem=$(free -g|awk '/Mem:/{print $2}')
    if (( mem < MIN_MEMORY_GB )); then
      record_fail "Memory" "${mem} GB"
    elif (( mem < RECOMMENDED_MEMORY_GB )); then
      record_warn "Memory" "${mem} GB"
    else
      record_pass "Memory" "${mem} GB"
    fi
}

doctor_check_disk(){
    free=$(df -BG "${LAB_ROOT}"|awk 'NR==2{gsub("G","",$4);print $4}')
    if (( free < MIN_DISK_GB )); then
      record_fail "Disk" "${free} GB free"
    elif (( free < RECOMMENDED_DISK_GB )); then
      record_warn "Disk" "${free} GB free"
    else
      record_pass "Disk" "${free} GB free"
    fi
}

doctor_execute(){
 section "System Validation"
 doctor_check_os
 doctor_check_cpu
 doctor_check_memory
 doctor_check_disk
}

doctor_summary(){
 section "Summary"
 summary_line Passed "$PASS_COUNT"
 summary_line Warnings "$WARN_COUNT"
 summary_line Failed "$FAIL_COUNT"
}

doctor_service(){
 section "Android Framework Research Doctor"
 doctor_initialize
 doctor_execute
 doctor_summary
}
