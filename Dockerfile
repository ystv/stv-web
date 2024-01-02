FROM golang:1.21.5-alpine3.19 AS build

LABEL site="ystv-stv-web"
LABEL stage="builder"

VOLUME /db
VOLUME /toml

WORKDIR /src/

COPY go.mod ./
COPY go.sum ./
COPY . ./
RUN go mod download

RUN apk update && apk add git && apk add make && apk add protoc && apk add protoc-gen-go --repository https://dl-cdn.alpinelinux.org/alpine/edge/testing/ --allow-untrusted

COPY *.go ./

RUN GOOS=linux GOARCH=amd64 make

EXPOSE 6691

ENTRYPOINT ["./stv-web"]