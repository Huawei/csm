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

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
PROG="${0}"
PUSH_IMAGE_OUTPUT="/tmp/huawei-csm-upload.$$.out"
VERSION={{version}}
IMAGE_LOCAL_PATH="/opt/huawei-csm/image/"
IMAGES=("csm-prometheus-collector" "csm-cmi" "csm-topo-service" "csm-liveness-probe")

# export the name of the debug log, so child processes will see it
export DEBUGLOG="${SCRIPTDIR}/upload-debug.log"

source "$SCRIPTDIR"/common.sh
if [ -f "${DEBUGLOG}" ]; then
  rm -f "${DEBUGLOG}"
fi

# usage will print command execution help and then exit
function usage() {
  decho
  decho "Help for $PROG"
  decho
  decho "Usage: $PROG options..."
  decho "Options:"
  decho "  Required"
  decho "  --imageRepo[=]<image repository>             Project image repository address, such as 'docker.io/library/'"

  exit 0
}

while getopts ":h-:" optchar; do
  case "${optchar}" in
  -)
    case "${OPTARG}" in
    # IMAGE REPO
    imageRepo)
      imageRepo="${!OPTIND}"
      OPTIND=$((OPTIND + 1))
      ;;
    imageRepo=*)
      imageRepo=${OPTARG#*=}
      ;;
    *)
      decho "Unknown option --${OPTARG}"
      decho "For help, run $PROG -h"
      exit 1
      ;;
    esac
    ;;
  h)
    usage
    ;;
  *)
    decho "Unknown option -${OPTARG}"
    decho "For help, run $PROG -h"
    exit 1
    ;;
  esac
done

# validate_params will validate the parameters passed in
function validate_params() {
  # IMAGE REPO
  if [ -z "${imageRepo}" ]; then
    decho "imageRepo must be specified"
    usage
    exit 1
  fi
}

# print header information
function header() {
  log section "Uploading Huawei-CSM Images to Image Repository"
}

# check docker command
function check_docker_command() {
  docker --help >&/dev/null || {
    decho "docker required for upload... exiting"
    exit 2
  }
}

# load all images locally
function load_images() {
  # shellcheck disable=SC2086
  for image in ${IMAGES[@]}; do
    run_command docker load -i ${IMAGE_LOCAL_PATH}${image}-${VERSION}.tar >/dev/null 2>&1
    if [ $? -ne 0 ]; then
      error=1
      log step_failure
      exit 2
    fi
  done
  log step_success
}

# tag all images
function tag_images() {
  for image in ${IMAGES[@]}; do
    run_command docker tag ${image}:${VERSION} ${imageRepo}${image}:${VERSION} >/dev/null 2>&1
    if [ $? -ne 0 ]; then
      error=1
      log step_failure
      exit 2
    fi
  done
  log step_success
}

# push all images
function push_images() {
  error=0
  # shellcheck disable=SC2086
  for image in ${IMAGES[@]}; do
    run_command docker push ${imageRepo}${image}:${VERSION} >"${PUSH_IMAGE_OUTPUT}" 2>&1
    if [ $? -ne 0 ]; then
      error=1
      log step_failure
      return 1
    fi
  done
  log step_success
  return 0
}

function cleanup_images() {
  # shellcheck disable=SC2086
  for image in ${IMAGES[@]}; do
    run_command docker rmi ${imageRepo}${image}:${VERSION} >/dev/null 2>&1
    # shellcheck disable=SC2181
    if [ $? -ne 0 ]; then
      error=1
      log step_failure
      exit 2
    fi
    run_command docker rmi ${image}:${VERSION} >/dev/null 2>&1
    if [ $? -ne 0 ]; then
      error=1
      log step_failure
      exit 2
    fi
  done
  log step_success
}

# upload the xuanwu images to image repository
function upload() {
  log step "Start to upload images"

  log step_success

  log arrow
  log smart_step "Start to load the images" "small"
  load_images

  log arrow
  log smart_step "Start to tag the images" "small"
  tag_images

  log arrow
  log smart_step "Start to push the images to the image repository" "small"
  push_images
  if [ $? -ne 0 ]; then
    cat "${PUSH_IMAGE_OUTPUT}"
    warning "Can not push the image to ${imageRepo}, plz check the connection of the image repository."
    cleanup_images
    exit 1
  fi

  log arrow
  log smart_step "Start to cleanup local images" "small"
  cleanup_images

  log arrow
  log smart_step "Images uploaded" "small"
  log step_success
}

# main used to upload the image step by step
function main() {
  validate_params
  check_docker_command
  header
  upload
}

main
