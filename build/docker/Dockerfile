# The build/ci environment for this project.
ARG BUILD_IMAGE=arm64v8/golang
FROM $BUILD_IMAGE:1.19.2-alpine3.16

RUN apk add --no-cache --upgrade \
    bash \
    curl \
    make \    
    grep \
    g++ \
    git \
    && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.1 \
    && go install github.com/matryer/moq@v0.3.0