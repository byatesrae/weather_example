#!/bin/bash

set -e

echo " * Linting (optional)..."
echo

git config --global --add safe.directory /src 

$(go env GOPATH)/bin/golangci-lint run -c ./.golangci.optional.yml

echo
echo " * Done."
