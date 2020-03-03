#!/bin/bash

# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# set -o errexit
set -o nounset
set -o pipefail

ROOT_PATH="$(dirname $(realpath $0))"/..

# create a temporary directory
TMP_DIR=$(mktemp -d)

# cleanup on exit
cleanup() {
  echo "Cleaning up..."
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

# install shellcheck (Linux-x64 only!)
cd "${TMP_DIR}" || exit

if [[ "$OSTYPE" == "msys" ]]; then
  DOWNLOAD_FILE="shellcheck-stable.zip"
  curl https://storage.googleapis.com/shellcheck/"${DOWNLOAD_FILE}" --output  "${DOWNLOAD_FILE}"
  unzip "${DOWNLOAD_FILE}"
  BIN_PATH="${TMP_DIR}/shellcheck-stable.exe"
else
  VERSION="shellcheck-v0.6.0"
  DOWNLOAD_FILE="${VERSION}.linux.x86_64.tar.xz"
  curl https://storage.googleapis.com/shellcheck/"${DOWNLOAD_FILE}" --output "${DOWNLOAD_FILE}"
  tar xf "${DOWNLOAD_FILE}"
  BIN_PATH="${TMP_DIR}/${VERSION}/shellscheck"
fi

echo "Running shellcheck..."
cd "${ROOT_PATH}" || exit
OUT="${TMP_DIR}/out.log"
FILES=$(ls -r | grep ".sh")
while read -r file; do
  "${BIN_PATH}" "$file" >> "${OUT}" 2>&1
done <<< "$FILES"

if [[ -s "${OUT}" ]]; then
  echo "Found errors:"
  cat "${OUT}"
  exit 1
fi
