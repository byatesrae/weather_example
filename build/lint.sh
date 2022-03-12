#!/bin/bash

set -e

# Run unit tests
echo " * Linting ..."
echo

$(go env GOPATH)/bin/golangci-lint run

echo
echo " * Done."
