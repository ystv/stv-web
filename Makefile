# parameters
GOBUILD=CGO_ENABLED=0 go build
GOCLEAN=go clean
PROTOC=protoc
BINARY_NAME=stv-web

PROTO_GENERATED=storage/storage.pb.go

.DEFAULT_GOAL := build

%.pb.go: %.proto
	$(PROTOC) -h
	$(PROTOC) -I=storage/ --go_opt=paths=source_relative --go_out=storage/ $<

build: $(PROTO_GENERATED)
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/stv-web
.PHONY: build

clean:
	$(GOCLEAN)
	rm -f $(PROTO_GENERATED)
.PHONY: clean

all: build
.PHONY: all
