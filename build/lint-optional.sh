#!/bin/bash

set -e

echo " * Linting (optional)..."
echo

$(go env GOPATH)/bin/golangci-lint run -c ./.golangci.optional.yml -n --new-from-rev master

echo
echo " * Done."
