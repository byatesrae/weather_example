FROM golang:1.19.0-alpine3.16

WORKDIR /opt/app

CMD ["go", "run", "/opt/app/cmd/weatherapi/"]
