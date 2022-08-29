#!/bin/bash

set -e

# Run unit tests
echo " * Generating test coverage ..."
echo

go test -json -covermode=atomic -coverprofile=./coverage.out ./... | tee ./gotest.json
go tool cover -html=coverage.out -o coverage.html

echo
echo " * Done."
