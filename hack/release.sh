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
set -o nounset
set -o pipefail

script_path=$(dirname "$(realpath "$0")")
pushd "$script_path/.."
mkdir -p ./_output
mkdir -p ./_output/assets

app_name=app
app_path=./"$app_name"/...

function build_os() {
  local os="$1"
  local ext="$2"
  shift 2
  local arches=("$@")
  for arch in "${arches[@]}"; do
    (set -x; GOOS="$os" GOARCH="$arch" go build -o ./_output/"$os/$arch/$app_name$ext" "$app_path")
    local asset="$app_name-$os-$arch.tar.gz"
    (set -x; tar -C ./_output/"$os/$arch" -czvf ./_output/assets/"$asset" "$app_name$ext" > /dev/null)
    (set -x; sha256sum ./_output/assets/"$asset" > ./_output/assets/"$asset.sha256")
  done
}

arch_linux=(amd64 arm arm64 ppc64 s390x)
arch_windows=(amd64)

build_os "linux"   ""     "${arch_linux[@]}"
build_os "windows" ".exe" "${arch_windows[@]}"
popd
