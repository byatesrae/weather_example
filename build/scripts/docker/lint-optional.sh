#!/bin/bash

# Runs ./build/scripts/lint-optional.sh dockerized.
# Intended to be invoked from the repository root.

set -e

source ./build/scripts/docker/common.sh

read_env_file

trap reset_owner_of_files ERR

echo " * Running lint-optional dockerised ..."
docker run \
    --rm \
    -v ${PWD}:/src \
    --env-file=.env \
    --workdir="/src" \
    --entrypoint /bin/bash \
    $BUILD_IMAGE \
    "-c" "./build/scripts/docker/config.sh; ./build/scripts/lint-optional.sh;" 

reset_owner_of_files