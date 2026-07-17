#!/usr/bin/env bash

command_bootstrap() {
    source "${LAB_ROOT}/lib/services/bootstrap.sh"

    bootstrap_service "$@"
}
