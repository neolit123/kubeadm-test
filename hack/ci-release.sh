#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
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

set -o errexit
set -o pipefail

# validation
if [[ -z "$TOKEN" ]]; then
  echo "error: the environment variable TOKEN must be set to a valid GitHub token!"
  exit 1
fi
if [[ -z "$DEST" ]]; then
  echo "error: the environment variable DEST must be set to a valid GitHub 'org/repo'!"
  exit 1
fi
if [[ -z "$RELEASE_TAG" ]]; then
  echo "error: the environment variable RELEASE_TAG must be set to a valid tag for '$DEST'!"
  exit 1
fi

SCRIPT_PATH=$(dirname "$(realpath "$0")")

# create a temporary directory
TMP_DIR=$(mktemp -d)
echo "created temporary directory: $TMP_DIR"

# cleanup
exitHandler() (
  echo "removing $TMP_DIR..."
  rm -rf "${TMP_DIR}"
)
trap exitHandler EXIT

# pull the kubernetes/release repository and build the release-notes tool
pushd "${TMP_DIR}"
(set -x; git clone --depth=1 --branch=v0.2.5 \
  https://github.com/kubernetes/release 2> /dev/null)
pushd ./release/cmd/release-notes
(set -x; go build)
RELEASE_NOTES_TOOL_PATH="$(realpath ./release-notes)"
echo "* using release notes tool path: $RELEASE_NOTES_TOOL_PATH"
popd
popd

# build all release artifacts
MAKEFILE_PATH="${SCRIPT_PATH}/../Makefile"
(set -x; make -f "${MAKEFILE_PATH}" clean)
(set -x; make -f "${MAKEFILE_PATH}" release)

# build the "create-release" tool
pushd "${SCRIPT_PATH}/../k8s-repo-tools/cmd/k8s-create-release"
CREATE_RELEASE_PATH="${TMP_DIR}/k8s-create-release"
(set -x; go build -o "${CREATE_RELEASE_PATH}")
popd

# prepare the asset flags
ASSETS_PATH="$(realpath "${SCRIPT_PATH}/../_output/assets")"
ASSETS=$(ls -1A "${ASSETS_PATH}")
ASSETS_FLAGS=""
while IFS= read -r asset; do
  ASSETS_FLAGS="$ASSETS_FLAGS -release-asset $asset=$ASSETS_PATH/$asset"
done <<< "$ASSETS"

# execute the "create-release" tool.
# "timeout" affects artifact upload timeout as well!
# shellcheck disable=SC2086
"${CREATE_RELEASE_PATH}" \
  -dry-run=false \
  -force \
  -token "${TOKEN}" \
  -dest "${DEST}" \
  -release-tag "${RELEASE_TAG}" \
  -release-notes-tool-path "${RELEASE_NOTES_TOOL_PATH}" \
  -timeout "5m" \
  ${ASSETS_FLAGS}
