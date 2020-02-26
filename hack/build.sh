#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
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

ROOT="$(dirname "$0")"
source "${ROOT}"/version.sh

pushd "${KUBE_ROOT}" > /dev/null

# set -o xtrace

latest_tag=$(git for-each-ref refs/tags --sort=-taggerdate --format='%(refname)' --count=1 | cut -d '/' -f 3)
commits_since_tag=$(git rev-list $latest_tag..HEAD --count)
sha=$(git rev-parse --short=14 HEAD)

#    KUBE_GIT_COMMIT - The git commit id corresponding to this
#          source code.
#    KUBE_GIT_TREE_STATE - "clean" indicates no changes since the git commit id
#        "dirty" indicates source code changes after the git commit id
#        "archive" indicates the tree was produced by 'git archive'
#    KUBE_GIT_VERSION - "vX.Y" used to indicate the last release version.
#    KUBE_GIT_MAJOR - The major part of the version
#    KUBE_GIT_MINOR - The minor component of the version


if [[ $commits_since_tag == "0" ]]; then 
  export KUBE_GIT_VERSION=$latest_tag
else
  export KUBE_GIT_VERSION=$latest_tag-$commits_since_tag-g$sha
fi
export KUBE_GIT_COMMIT=$(git rev-parse HEAD)

LDFLAGS=$(kube::version::ldflags)

COMMAND="go build -ldflags \"${LDFLAGS}\""
echo "${COMMAND}"
eval "${COMMAND}"



# tag count
# git rev-list v1.19.1..HEAD --count

# latest tag
# git for-each-ref refs/tags --sort=-taggerdate --format='%(refname)' --count=1

# v1.19.0-1-g2fef93fc41697a

 
