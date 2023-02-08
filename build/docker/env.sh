#!/bin/bash

# Runs ./build/env.sh dockerized.
# Intended to be invoked from the repository root.

set -e

source ./build/docker/common.sh

read_env_file

trap reset_owner_of_files ERR

echo " * Running env dockerised ..."
docker run \
    --rm \
    -v ${PWD}:/src \
    --env-file=.env \
    --workdir="/src" \
    --entrypoint /bin/bash \
    $BUILD_IMAGE \
    "-c" "./build/docker/config.sh; ./build/env.sh;" 

reset_owner_of_files