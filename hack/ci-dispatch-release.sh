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

# This script can be executed to dispatch an event to a repository to create a release.
# Normally a workflow can listen for the creation of new tags, expect that for some
# reason this does not trigger when the tag is created by a GitHub workflow.

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
if [[ -z "$SYNC_OUTPUT" ]]; then
  echo "error: the environment variable SYNC_OUTPUT must be set to a repo-sync output file!"
  exit 1
fi

set -x
curl --version
REFS=$(jq -c ".[]" < "$SYNC_OUTPUT")
EVENT_TYPE="dispatch-k8s-create-release"
RELEASE_TAG="release_tag"
set +x

while IFS= read -r ref; do
  # ignore refs non-tag refs
  if [[ $ref != *"refs/tags"* ]]; then
    continue
  fi

  # get the ref value
  ref=$(echo "$ref" | jq -r ".ref")
  echo "* sending request to $DEST to create release from ref $ref"

  # send a POST request to the destination repository that contains
  # the tag ref as payload and matches the event_type that
  # the create-release workflow triggers on.
  curl -H "Accept: application/vnd.github.everest-preview+json" \
    -H "Authorization: token $TOKEN" \
    --request POST \
    --data "{\"event_type\": \"$EVENT_TYPE\", \"client_payload\": {\"$RELEASE_TAG\": \"$ref\"}}" \
    https://api.github.com/repos/"$DEST"/dispatches
done <<< "$REFS"
