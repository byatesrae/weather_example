#!/bin/bash

# Runs ./build/scripts/lint.sh dockerized.
# Intended to be invoked from the repository root.

set -e

source ./build/scripts/docker/common.sh

read_env_file

trap reset_owner_of_files ERR

# TODO - Mount caches with something like...
# -v $(go env GOMODCACHE):/go/pkg/mod \
# -v $(go env GOCACHE):/root/.cache/go-build \
# -e GOLANGCI_LINT_CACHE=/root/.cache/go-build \
echo " * Running lint dockerised ..."
docker run \
    --rm \
    -v ${PWD}:/src \
    --env-file=.env \
    --workdir="/src" \
    --entrypoint /bin/bash \
    $BUILD_IMAGE \
    "-c" "./build/scripts/docker/config.sh; ./build/scripts/lint.sh;" 

reset_owner_of_files