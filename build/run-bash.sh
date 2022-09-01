#!/bin/bash

set -e

reset_permissions() {
    docker run --rm -v $(pwd):/src busybox:stable chown -R $(id -u):$(id -u) src
}

trap reset_permissions ERR

if [ -z "$1" ]; then 
    echo "ERR: First argument must be command to run.";
    exit 1; 
fi

docker build -t weather_example_build ./build/

docker run \
    --rm \
    -v ${PWD}:/src \
    --env-file=.env \
    --workdir="/src" \
    --entrypoint /bin/bash \
    weather_example_build \
    "-c" "chown -R root:root /src && $1" # Not ideal using root here

reset_permissions
