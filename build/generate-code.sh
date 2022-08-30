#!/bin/bash

set -e

echo " * Generating code ..."
echo

go generate ./...

echo
echo " * Done."
