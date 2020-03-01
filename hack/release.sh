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

script_path=$(dirname "$(realpath "$0")")
# shellcheck disable=SC1090
source "$script_path"/version.sh

# If KUBE_GIT_VERSION is not set, attempt to set it outside of the kube::version::ldflags logic.
# This workarounds the issue around "git describe" not using the latest tag.
if [[ -z "$KUBE_GIT_VERSION" ]]; then

  # Check if running in a git repository
  if [[ $(git rev-parse --is-inside-work-tree 2> /dev/null) == "true" ]]; then
    echo "* running in a git repository"

    # Get the latest tag based on the current branch
    git_branch=$(git rev-parse --abbrev-ref HEAD)
    if [[ $git_branch == "master" ]]; then
      # If on the master branch get the newest tag
      latest_tag=$(git for-each-ref refs/tags --sort=-taggerdate --format='%(refname)' --count=1 | cut -d '/' -f 3)
    else
      # If on a release branch get the newest tag for this release
      git_branch=${git_branch#"release-"}
      latest_tag=$(git for-each-ref refs/tags --sort=-taggerdate --format='%(refname)' | cut -d '/' -f 3 | grep "$git_branch" | head -n 1)
    fi

    # Count the number of commits since the latest tag and get the short SHA
    commits_since_tag=$(git rev-list "$latest_tag"..HEAD --count)
    sha=$(git rev-parse --short=14 HEAD)

    # Format the KUBE_GIT_VERSION
    if [[ $commits_since_tag == "0" ]]; then
      export KUBE_GIT_VERSION=$latest_tag
    else
      export KUBE_GIT_VERSION=$latest_tag-$commits_since_tag-g$sha
    fi

  else
    # If this is not a repository require KUBE_GIT_VERSION_FILE to be set
    echo "* not running in a git repository"

    if [[ -z "$KUBE_GIT_VERSION_FILE" ]]; then
      echo "error: KUBE_GIT_VERSION_FILE must be set. See ./hack/version.sh for details."
      exit 1
    fi
  fi
fi

# Generate ldflags
export KUBE_ROOT="$script_path"/..
ldflags=$(kube::version::ldflags)

pushd "$KUBE_ROOT"
mkdir -p ./_output
mkdir -p ./_output/assets

app_name=app
app_path=./"$app_name"

# Build all architectures for a given OS
function build_os() {
  local os="$1"
  local ext="$2"
  shift 2
  local arches=("$@")
  for arch in "${arches[@]}"; do
    pushd "$app_path" > /dev/null
    (set -x; GOOS="$os" GOARCH="$arch" go build -o ../_output/"$os/$arch/$app_name$ext" -ldflags "$ldflags")
    popd > /dev/null
    local asset="$app_name-$os-$arch.tar.gz"
    (set -x; tar -C ./_output/"$os/$arch" -czvf ./_output/assets/"$asset" "$app_name$ext" > /dev/null)
    (set -x; sha256sum ./_output/assets/"$asset" > ./_output/assets/"$asset.sha256")
  done
}

arch_linux=(amd64 arm arm64 ppc64 s390x)
arch_windows=(amd64)

build_os "linux"   ""     "${arch_linux[@]}"
build_os "windows" ".exe" "${arch_windows[@]}"
