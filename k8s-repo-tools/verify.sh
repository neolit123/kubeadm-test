#!/bin/bash
set -o xtrace

go vet ./... || exit 1
golint ./... || exit 1
go test ./... || exit 1
staticcheck ./... || exit 1
