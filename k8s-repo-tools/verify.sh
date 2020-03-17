#!/bin/bash
set -o xtrace

test -z "$(gofmt -d ./)" || exit 1
go vet ./... || exit 1
golint ./... || exit 1
go test -count=1 ./... || exit 1
staticcheck ./... || exit 1
