FROM golang:1.23.3-alpine3.20 AS build

LABEL site="ystv-stv-web"
LABEL stage="builder"
LABEL author="Liam Burnand"

ARG STV_WEB_VERSION_ARG
ARG STV_WEB_COMMIT_ARG

VOLUME /db
VOLUME /toml

WORKDIR /src/

COPY go.mod ./
COPY go.sum ./
COPY . ./
RUN go mod download

RUN apk update
RUN apk add git make protoc
RUN apk add protoc-gen-go --repository https://dl-cdn.alpinelinux.org/alpine/edge/testing/ --allow-untrusted

COPY *.go ./

# Set build variables
RUN echo -n "-X 'main.Version=$STV_WEB_VERSION_ARG" > ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "' -X 'main.Commit=$STV_WEB_COMMIT_ARG" >> ./ldflags && \
    tr -d \\n < ./ldflags > ./temp && mv ./temp ./ldflags && \
    echo -n "'" >> ./ldflags

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make LDFLAGS="$(cat ./ldflags)"

EXPOSE 6691

ENTRYPOINT ["./stv-web"]