#!/usr/bin/env bash

# Runs pkgsite for hosting source code documentation. See https://tip.golang.org/doc/comment.
# Intended to be invoked from the repository root.

set -e

echo " * Hosting documentation ..."
echo

if [ ! -z $1 ]
then
    DOC_PORT=$1
else
    DOC_PORT=6060
fi

if [ ! -f "go.mod" ]; then
    echo "go.mod does not exist"
    exit 1
fi

MODULE=$(cat go.mod | grep -oP '^module\s+\K\S+')

echo " * Visit http://localhost:$DOC_PORT/$MODULE"

pkgsite -http :$DOC_PORT

echo
echo " * Done."