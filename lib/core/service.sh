#!/usr/bin/env bash

# ============================================================
# ASRL Service Framework
#
# Every service follows the lifecycle:
#
# initialize()
# execute()
# summary()
# cleanup()
#
# ============================================================

service_initialize() {
    :
}

service_execute() {
    :
}

service_summary() {
    :
}

service_cleanup() {
    :
}

service_run() {

    service_initialize

    service_execute

    service_summary

    service_cleanup
}