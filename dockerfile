ARG BUILD_IMAGE=coverto/go_build
FROM $BUILD_IMAGE As Build

RUN mkdir /src
WORKDIR /src

COPY . .

RUN go build -o ./bin/app ./cmd/weatherapi/

FROM alpine:3.17.2

WORKDIR /opt/app

COPY --from=Build /src/bin .

# Add a non-root user
RUN adduser -D appuser
USER appuser

ENTRYPOINT ["./app"]
