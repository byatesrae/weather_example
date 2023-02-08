#!/bin/bash

# Runs the docs command in a dockerised environment.
# Intended to be invoked from the repository root. For example:
#
# ./build/docker/run.sh "make env lint; echo Done!"

set -e

if [ ! -f  ".env" ]; then
    echo "ERR: \".env\" file not found."
    
    exit 1; 
fi

source .env

# TODO: Instead of re-building this every time we could build once and host the 
# image(s) for pull.
echo " * Building image ..."
docker build \
    -t weather_example_docs \
    $( [ -n "$BUILD_IMAGE" ] && printf %s "--build-arg BUILD_IMAGE=$BUILD_IMAGE" ) \
    ./build/docker/

# Run the doc command.
echo " * Running \"$1\" ..."
docker run \
    --rm \
    -v ${PWD}:/src \
    --env-file=.env \
    --workdir="/src" \
    --entrypoint /bin/bash \
    -p $DOC_PORT:$DOC_PORT \
    weather_example_docs \
    "-c" "./docs/docs.sh" 