#!/usr/bin/env bash

set -euo pipefail

PROJECT_NAME="Android Security Research Lab"

VERSION_FILE="$(dirname "$0")/../VERSION"

show_banner() {
    echo "=============================================="
    echo "        ${PROJECT_NAME}"
    echo "=============================================="
}

show_help() {

cat <<EOF

Usage:

    lab <command>

Commands

    help            Show help

    version         Show version

    doctor          Check environment

    bootstrap       Initialize Ubuntu

    workspace       Manage workspace

    repo            Manage repo

    build           Build Android

    research        Research management

EOF

}

show_version() {

if [[ -f "${VERSION_FILE}" ]]; then

    VERSION=$(cat "${VERSION_FILE}")

else

    VERSION="Unknown"

fi

echo "${PROJECT_NAME}"

echo

echo "Version : ${VERSION}"

}

case "${1:-help}" in

help)

    show_banner

    show_help

;;

version)

    show_version

;;

doctor)

    echo "Not Implemented Yet"

;;

bootstrap)

    echo "Not Implemented Yet"

;;

workspace)

    echo "Not Implemented Yet"

;;

repo)

    echo "Not Implemented Yet"

;;

build)

    echo "Not Implemented Yet"

;;

research)

    echo "Not Implemented Yet"

;;

*)

    echo "Unknown Command"

    echo

    show_help

    exit 1

;;

esac