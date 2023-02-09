#!/bin/bash

# Runs ./build/scripts/docs.sh dockerized.
# Intended to be invoked from the repository root.

set -e

source ./build/scripts/docker/common.sh

read_env_file

echo " * Running docs dockerised ..."
docker run \
    --rm \
    -v ${PWD}:/src \
    --env-file=.env \
    --workdir="/src" \
    --entrypoint /bin/bash \
    -p $DOCS_PORT:$DOCS_PORT \
    $BUILD_IMAGE \
    "-c" "./build/scripts/docker/config.sh; ./build/scripts/docs.sh;" 