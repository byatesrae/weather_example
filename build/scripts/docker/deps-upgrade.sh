#!/bin/bash

# Runs ./build/scripts/deps-upgrade.sh dockerized.
# Intended to be invoked from the repository root.

set -e

source ./build/scripts/docker/common.sh

read_env_file

trap reset_owner_of_files ERR

echo " * Running deps-upgrade dockerised ..."
docker run \
    --rm \
    -v ${PWD}:/src \
    --env-file=.env \
    --workdir="/src" \
    --entrypoint /bin/bash \
    $BUILD_IMAGE \
    "-c" "./build/scripts/docker/config.sh; ./build/scripts/deps-upgrade.sh;" 

reset_owner_of_files