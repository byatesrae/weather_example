#!/usr/bin/env bash

# Runs pkgsite for hosting source code documentation. See https://tip.golang.org/doc/comment.
# Intended to be invoked from the repository root.

set -e

echo " * Hosting documentation ..."
echo

if [ -z "$DOCS_PORT" ]; then
    echo "ERR: environment variable DOCS_PORT is required."        
    exit 1
fi

if [ ! -f "go.mod" ]; then
    echo "go.mod does not exist"
    exit 1
fi

MODULE=$(cat go.mod | grep -oP '^module\s+\K\S+')

echo " * Visit http://localhost:$DOCS_PORT/$MODULE once server has started."

pkgsite -http :$DOCS_PORT

echo
echo " * Done."