#!/bin/bash

set -e

# Run unit tests
echo " * Cleaning ..."
echo

rm -rf coverage.out coverage.html golangci.out gotest.json

echo
echo " * Done."
