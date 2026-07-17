#!/usr/bin/env bash

# ============================================================
# Global Constants
# ============================================================

readonly PROJECT_NAME="Android Security Research Lab"
readonly PROJECT_SHORT="ASRL"

readonly VERSION_FILE="${LAB_ROOT}/VERSION"

readonly PROJECT_VERSION_FILE="${LAB_ROOT}/VERSION"

readonly CONFIG_DIR="${LAB_ROOT}/config"
readonly LAB_CONFIG_TEMPLATE="${CONFIG_DIR}/lab.conf.example"
readonly LAB_CONFIG_FILE="${LAB_CONFIG_FILE:-${CONFIG_DIR}/lab.conf}"

readonly OUTPUT_DIR="${LAB_ROOT}/output"

readonly DOCTOR_CONFIG="${LAB_ROOT}/config/doctor.conf"
readonly BOOTSTRAP_CONFIG="${LAB_ROOT}/config/bootstrap.conf"
