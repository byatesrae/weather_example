#!/bin/bash

set -e

echo " * Running tests ..."
echo

go test -race ./...

echo
echo " * Done."
