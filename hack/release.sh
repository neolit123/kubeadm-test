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

set -o xtrace
set -o errexit
set -o nounset
set -o pipefail

SCRIPT_PATH=$(dirname "$(realpath "$0")")

pushd "$SCRIPT_PATH/.."
mkdir -p ./_output
GOOS=windows GOARCH=amd64 go build -o ./_output/windows/amd64/app.exe ./app/...
GOOS=linux GOARCH=amd64   go build -o ./_output/linux/amd64/app       ./app/...
GOOS=linux GOARCH=arm     go build -o ./_output/linux/arm/app         ./app/...
GOOS=linux GOARCH=arm64   go build -o ./_output/linux/arm64/app       ./app/...
GOOS=linux GOARCH=ppc64   go build -o ./_output/linux/ppc64/app       ./app/...
GOOS=linux GOARCH=s390x   go build -o ./_output/linux/s390x/app       ./app/...
popd
