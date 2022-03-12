FROM golang:1.17.3-alpine3.14

WORKDIR /opt/app

CMD ["go", "run", "/opt/app/cmd/weatherapi/"]
