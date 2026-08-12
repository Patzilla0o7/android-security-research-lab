#!/usr/bin/env bash

# ============================================================
# Global Constants
# ============================================================

readonly PROJECT_NAME="Android Security Research Lab"
readonly PROJECT_SHORT="ASRL"

readonly VERSION_FILE="${LAB_ROOT}/VERSION"

readonly PROJECT_VERSION_FILE="${LAB_ROOT}/VERSION"

readonly EXIT_SUCCESS=0
readonly EXIT_FAILURE=1
readonly EXIT_USAGE=2
readonly EXIT_NOT_IMPLEMENTED=3

readonly CONFIG_DIR="${LAB_ROOT}/config"
readonly LAB_CONFIG_TEMPLATE="${CONFIG_DIR}/lab.conf.example"
readonly LAB_CONFIG_FILE="${LAB_CONFIG_FILE:-${CONFIG_DIR}/lab.conf}"

readonly OUTPUT_DIR="${OUTPUT_DIR:-${LAB_ROOT}/output}"

readonly DOCTOR_CONFIG="${LAB_ROOT}/config/doctor.conf"
readonly TOOLCHAIN_CONFIG="${LAB_ROOT}/config/tools.conf"
