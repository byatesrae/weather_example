#!/bin/bash

set -e

# Run unit tests
echo " * Running tests ..."
echo

go test -race ./...

echo
echo " * Done."
