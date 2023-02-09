#!/bin/bash

# Runs linting with more linters that are not required for CI to pass.
# Intended to be invoked from the repository root.

set -e

echo " * Linting (optional)..."
echo

$(go env GOPATH)/bin/golangci-lint run -c ./build/.golangci.optional.yml -n

echo
echo " * Done."
