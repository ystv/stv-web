# parameters
GO_BUILD=go build
GO_CLEAN=go clean
PROTOC=protoc
BINARY_NAME=stv-web

PROTO_GENERATED=storage/storage.pb.go

.DEFAULT_GOAL := build

%.pb.go: %.proto
	$(PROTOC) -I=storage/ --go_opt=paths=source_relative --go_out=storage/ $<

build: $(PROTO_GENERATED)
	$(GO_BUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)
.PHONY: build

clean:
	$(GO_CLEAN)
	rm -f $(PROTO_GENERATED)
.PHONY: clean

all: build
.PHONY: all
