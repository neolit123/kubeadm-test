#!/bin/bash

SCRIPT_PATH=$(dirname "$(realpath "$0")")

set -o xtrace

pushd "${SCRIPT_PATH}/../../k8s-repo-tools"
test -z "$(gofmt -d ./)" || exit 1
go vet ./... || exit 1
golint ./... || exit 1
go test -count=1 ./... || exit 1
staticcheck ./... || exit 1
