# Go Build
This is a non-standalone component that drives local development and continuous
integration for applications written in Go.

This component only serves it's purpose when used as a Git Submodule:
1. In a repository for one or more applications written in Go.
1. Under a folder named "build".

## Requires
* Dependencies outlined in [the Go Build dockerfile](https://github.com/byatesrae/docker.go_build/blob/v1.2.0/Dockerfile).
* Docker (tested with 20.10.12).

## Setup
1. In the parent repository, add this repository as a submodule with:
    ```
    git submodule add https://github.com/byatesrae/go_build build/
    ````
1. Copy the contents of `./examples` to the parent repository. All files copied can
be extended.

## Development
Run the following in the parent repository:
```
make help
```