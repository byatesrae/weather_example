#!/bin/bash

set -e

echo " * Upgrading/Installing dependencies ..."
echo

go get -u -t -d -v ./...
go mod tidy
go mod vendor

echo
echo " * Done."
