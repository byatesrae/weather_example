#!/bin/bash

set -e

echo " * Linting ..."
echo

$(go env GOPATH)/bin/golangci-lint run

echo
echo " * Done."
