FROM golang:1.12.6-alpine3.9 AS build

RUN apk update && apk add git

WORKDIR /src/poc-go-nats

COPY . .

RUN go install .

FROM alpine:3.9

COPY --from=build /go/bin/poc-go-nats /usr/local/bin/poc-go-nats

ENTRYPOINT [ "poc-go-nats" ]
