#!/bin/bash

# Runs bash commands in the dockerised build/ci environment.
# Intended to be invoked from the repository root. For example:
#
# ./build/docker/run.sh "make env lint; echo Done!"

set -e

if [ -z "$1" ]; then 
    echo "ERR: First argument must be a bash command(s) to run. For example:
    ./build/docker/run.sh \"make env lint; echo Done!\"";
    
    exit 1; 
fi

if [ ! -f  ".env" ]; then
    echo "ERR: \".env\" file not found."
    
    exit 1; 
fi

# The user inside the docker container is root. Any files created need to have 
# permissions updated after the container exits.
reset_permissions() {
    echo " * Resetting permissions ..."
    docker run --rm -v $(pwd):/src busybox:stable chown -R $(id -u):$(id -u) src
}
trap reset_permissions ERR

source .env

# Run the bash command(s).
# TODO - Mount caches with something like...
# -v $(go env GOMODCACHE):/go/pkg/mod \
# -v $(go env GOCACHE):/root/.cache/go-build \
# -e GOLANGCI_LINT_CACHE=/root/.cache/go-build \
echo " * Running \"$1\" ..."
docker run \
    --rm \
    -v ${PWD}:/src \
    --env-file=.env \
    --workdir="/src" \
    --entrypoint /bin/bash \
    coverto/go_build:v1.1.0 \
    "-c" "./build/docker/config.sh; $1" 

reset_permissions