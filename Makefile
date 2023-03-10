# parameters
GOBUILD=CGO_ENABLED=0 go build
GOCLEAN=go clean
PROTOC=protoc
STATIK=statik
BINARY_NAME=stv-web

PROTO_GENERATED=storage/storage.pb.go
STATIK_GENERATED=statik/statik.go
PUBLIC_FILES=$(wildcard public/*)

.DEFAULT_GOAL := build

%.pb.go: %.proto
	$(PROTOC) -I=storage/ --go_opt=paths=source_relative --go_out=storage/ $<

#$(STATIK_GENERATED): $(PUBLIC_FILES)
#	echo "$(PUBLIC_FILES)"
#	$(STATIK) -f -src=templates/*.tmpl -dest=.

build: $(PROTO_GENERATED) $(STATIK_GENERATED)
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/stv-web
.PHONY: build

clean:
	$(GOCLEAN)
	rm -f $(PROTO_GENERATED)
	rm -f $(STATIK_GENERATED)
.PHONY: clean

all: build
.PHONY: all
