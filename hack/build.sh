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

# Require the APP_NAME environment variable to be set
if [[ -z "$APP_NAME" ]]; then
    echo "error: the env variable APP_NAME must be set"
    exit 1
fi
if [[ -z "$APP_PATH" ]]; then
    APP_PATH=./"$APP_NAME"
fi

# If KUBE_GIT_VERSION is not set, attempt to set it outside of the kube::version::ldflags logic.
# This workarounds the issue around "git describe" not using the latest tag.
if [[ -z "$KUBE_GIT_VERSION" ]]; then

  # Check if running in a git repository
  if [[ $(git rev-parse --is-inside-work-tree 2> /dev/null) == "true" ]]; then
    echo "* running in a git repository"

    # Get the current branch
    git_branch=$(git rev-parse --abbrev-ref HEAD)
    if [[ $git_branch == "master" ]]; then
      branch_flag=""
    else
      branch_flag=$git_branch
    fi

    # Use the k8s-latest-version tool to find the latest tag for this branch
    echo "* obtaining the latest tag for branch '$git_branch'"
    pushd "$script_path"/../k8s-repo-tools/cmd/k8s-latest-version > /dev/null
    latest_tag=$(git tag | go run main.go --branch="$branch_flag")
    popd > /dev/null
    if [[ -z "$latest_tag" ]]; then
      echo "error: could not obtain the latest tag using k8s-latest-version"
      exit 1
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

# Build architectures for a given OS and optionally create tarballs with a SHA256 checksum
function build_os() {
  local os="$1"
  local tarball=$2
  shift 2

  local arches=("$@")
  local ext=""
  if [[ "$os" == "windows" ]]; then
    ext=".exe"
  fi

  for arch in "${arches[@]}"; do
    pushd "$APP_PATH" > /dev/null
    (set -x; GOOS="$os" GOARCH="$arch" go build -o ../_output/"$os/$arch/$APP_NAME$ext" -ldflags "$ldflags")
    popd > /dev/null

    if [[ $tarball == true ]]; then
        local asset="$APP_NAME-$os-$arch.tar.gz"
        (set -x; tar -C ./_output/"$os/$arch" -czvf ./_output/assets/"$asset" "$APP_NAME$ext" > /dev/null)
        (set -x; sha256sum ./_output/assets/"$asset" > ./_output/assets/"$asset.sha256")
    fi
  done
}

# Build the default OS/ARCH
function build_os_default() {
    build_os "$(go env GOOS)" false "$(go env GOARCH)"
}

# Make it possible to call a function by sourcing this script
"$@"
