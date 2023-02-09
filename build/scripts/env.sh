#!/bin/bash

# Creates a default .env file.
# Intended to be invoked from the repository root.

set -e

if [ ! -f  ".env" ]; then
    echo " * Creating default .env file ..."
    echo    
    echo " * Don't forget to replace the \"FILL_IN_MANUALLY\" values in your new .env file."

    cp -n ./.env.local.example ./.env

    echo
    echo " * Done."
fi

