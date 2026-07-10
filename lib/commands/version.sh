#!/usr/bin/env bash

command_version() {

local version="Unknown"

if [[ -f "${VERSION_FILE}" ]]; then
    version=$(<"${VERSION_FILE}")
fi

echo "${PROJECT_NAME}"

echo

echo "Version : ${version}"

}