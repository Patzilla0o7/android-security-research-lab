#!/usr/bin/env bash

command_exists() {
    command -v "$1" >/dev/null 2>&1
}


require_command() {

    command -v "$1" >/dev/null 2>&1

}