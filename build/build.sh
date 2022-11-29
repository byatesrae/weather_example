#!/bin/bash

set -e

# Run linting
echo " * Building ${TARGET_APP}..."
echo

## TODO - swap out file extension based on OS
go build -o ./bin/${TARGET_APP}-${GOOS}-${GOARCH}$([ "$GOOS" == "windows" ] && echo ".exe") ./cmd/${TARGET_APP}/

echo
echo " * Done."
