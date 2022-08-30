#!/bin/bash

set -e

echo " * Installing dependencies ..."
echo

go mod tidy
go mod vendor

echo
echo " * Done."
