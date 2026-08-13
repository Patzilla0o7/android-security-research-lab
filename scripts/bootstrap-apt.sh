#!/usr/bin/env bash

# Minimal privileged adapter. Planning and package selection belong to Go.
set -euo pipefail

if (( $# == 0 )); then
    echo "No apt packages requested."
    exit 0
fi

sudo apt-get update
sudo apt-get install -y "$@"
