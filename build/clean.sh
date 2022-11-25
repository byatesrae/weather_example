#!/bin/bash

# Removes build artifacts and vendor directories.
# Intended to be invoked from the repository root.

set -e

echo " * Cleaning ..."
echo

rm -rf coverage.out coverage.html golangci.out gotest.json

echo
echo " * Done."
