#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#      http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# shellcheck disable=SC2034
DRIVERDIR="${SCRIPTDIR}/../helm"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
DARK_GRAY='\033[1;30m'
NC='\033[0m' # No Color

function decho() {
  if [ -n "${DEBUGLOG}" ]; then
    echo "$@" | tee -a "${DEBUGLOG}"
  fi
}

function debuglog_only() {
  if [ -n "${DEBUGLOG}" ]; then
    echo "$@" >> "${DEBUGLOG}"
  fi
}

function log() {
  case $1 in
  separator)
    decho "------------------------------------------------------"
    ;;
  error)
    decho
    log separator
    # shellcheck disable=SC2059
    printf "${RED}Error: $2\n"
    # shellcheck disable=SC2059
    printf "${RED}Installation cannot continue${NC}\n"
    debuglog_only "Error: $2"
    debuglog_only "Installation cannot continue"
    exit 1
    ;;
  uninstall_error)
    log separator
    # shellcheck disable=SC2059
    printf "${RED}Error: $2\n"
    # shellcheck disable=SC2059
    printf "${RED}Uninstallation cannot continue${NC}\n"
    debuglog_only "Error: $2"
    debuglog_only "Uninstallation cannot continue"
    exit 1
    ;;
  step)
    printf "|\n|- %-65s" "$2"
    debuglog_only "${2}"
    ;;
  small_step)
    printf "%-61s" "$2"
    debuglog_only "${2}"
    ;;
  section)
    log separator
    printf "> %s\n" "$2"
    debuglog_only "${2}"
    log separator
    ;;
  smart_step)
    if [[ $3 == "small" ]]; then
      log small_step "$2"
    else
      log step "$2"
    fi
    ;;
  arrow)
    printf "  %s\n  %s" "|" "|--> "
    ;;
  step_success)
    # shellcheck disable=SC2059
    printf "${GREEN}Success${NC}\n"
    ;;
  step_failure)
    # shellcheck disable=SC2059
    printf "${RED}Failed${NC}\n"
    ;;
  step_warning)
    # shellcheck disable=SC2059
    printf "${YELLOW}Warning${NC}\n"
    ;;
  info)
    printf "${DARK_GRAY}%s${NC}\n" "$2"
    ;;
  passed)
    # shellcheck disable=SC2059
    printf "${GREEN}Success${NC}\n"
    ;;
  warnings)
    # shellcheck disable=SC2059
    printf "${YELLOW}Warnings:${NC}\n"
    ;;
  errors)
    # shellcheck disable=SC2059
    printf "${RED}Errors:${NC}\n"
    ;;
  *)
    echo -n "Unknown"
    ;;
  esac
}

function run_command() {
  local RC=0
  if [ -n "${DEBUGLOG}" ]; then
    # shellcheck disable=SC2155
    local ME=$(basename "${0}")
    echo "---------------" >> "${DEBUGLOG}"
    # shellcheck disable=SC2145
    echo "${ME}:${BASH_LINENO[0]} - Running command: $@" >> "${DEBUGLOG}"
    debuglog_only "Results:"
    eval "$@" 2>&1 | tee -a "${DEBUGLOG}"
    RC=${PIPESTATUS[0]}
    echo "---------------" >> "${DEBUGLOG}"
  else
    eval "$@"
    RC=$?
  fi
  return $RC
}

# warning, with an option for users to continue
function warning() {
  log separator
  # shellcheck disable=SC2059
  printf "${YELLOW}WARNING:${NC}\n"
  for N in "$@"; do
    decho "$N"
  done
  decho
}
