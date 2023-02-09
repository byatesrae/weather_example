#!/bin/bash

# Configures the dockerized environment.

set -e

echo " * Configuring env ..."
echo

if [ -z "$GITHUB_TOKEN" ]; then 
    echo "ERR: GITHUB_TOKEN is required.";
    exit 1; 
fi

go env -w GOPRIVATE=github.com/byatesrae/*

git config --global url.https://x-access-token:$GITHUB_TOKEN@github.com/.insteadOf https://github.com/
git config --global --add safe.directory /src

echo
echo " * Done."