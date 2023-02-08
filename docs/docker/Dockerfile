# The build/ci environment for this project.
ARG BUILD_IMAGE=arm64v8/golang
FROM $BUILD_IMAGE:1.19.2-alpine3.16

RUN apk add --no-cache --upgrade \
    bash \
    && go install golang.org/x/pkgsite/cmd/pkgsite@latest