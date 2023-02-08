ARG BUILD_IMAGE=coverto/go_build
FROM $BUILD_IMAGE

WORKDIR /opt/app

CMD ["go", "run", "/opt/app/cmd/weatherapi/"]
