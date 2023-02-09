#!/bin/bash

# Runs all quality checks.
# Intended to be invoked from the repository root.

set -e

echo " * Running quality checks ..."
echo

./build/scripts/env.sh
./build/scripts/clean.sh
./build/scripts/lint.sh
./build/scripts/test.sh
./build/scripts/generate-test-coverage.sh