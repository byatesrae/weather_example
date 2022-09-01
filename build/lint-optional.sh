#!/bin/bash

set -e

echo " * Linting (optional)..."
echo

$(go env GOPATH)/bin/golangci-lint run -c ./.golangci.optional.yml

echo
echo " * Done."
