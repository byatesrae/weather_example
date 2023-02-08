#!/bin/bash

# Contains common logic shared among scripts.

set -e

# Reads in the .env file.
read_env_file() {
    if [ ! -f  ".env" ]; then
        echo "ERR: \".env\" file not found."
        
        exit 1
    fi

    source .env
}

# Will reset the owner of files created by root in the docker container used for these commands.
reset_owner_of_files() {
    echo " * Resetting owner of files ..."
    docker run --rm -v $(pwd):/src busybox:stable chown -R $(id -u):$(id -u) src
}