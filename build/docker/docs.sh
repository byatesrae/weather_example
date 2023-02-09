#!/bin/bash

# Runs ./build/docs.sh dockerized.
# Intended to be invoked from the repository root.

set -e

source ./build/docker/common.sh

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
    "-c" "./build/docker/config.sh; ./build/docs.sh;" 