#!/bin/bash

# Runs all quality checks.
# Intended to be invoked from the repository root.

set -e

echo " * Running quality checks ..."
echo

./build/env.sh
./build/clean.sh
./build/lint.sh
./build/test.sh
./build/generate-test-coverage.sh