#!/usr/bin/env bash
readonly COLOR_RESET="\033[0m"
readonly COLOR_RED="\033[31m"
readonly COLOR_GREEN="\033[32m"
readonly COLOR_YELLOW="\033[33m"
readonly COLOR_BLUE="\033[34m"

section(){ echo; echo "============================================================"; echo "$1"; echo "============================================================"; }
info(){ printf "${COLOR_BLUE}[INFO]${COLOR_RESET} %s\n" "$*"; }
success(){ printf "${COLOR_GREEN}[ OK ]${COLOR_RESET} %s\n" "$*"; }
warn(){ printf "${COLOR_YELLOW}[WARN]${COLOR_RESET} %s\n" "$*"; }
error(){ printf "${COLOR_RED}[FAIL]${COLOR_RESET} %s\n" "$*" >&2; }
doctor_pass(){ printf "${COLOR_GREEN}✓${COLOR_RESET} %-18s %s\n" "$1" "$2"; }
doctor_warn(){ printf "${COLOR_YELLOW}!${COLOR_RESET} %-18s %s\n" "$1" "$2"; }
doctor_fail(){ printf "${COLOR_RED}✗${COLOR_RESET} %-18s %s\n" "$1" "$2"; }
summary_line(){ printf "%-12s %s\n" "$1" "$2"; }
