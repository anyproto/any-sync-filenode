.PHONY: build test deps build-dev
SHELL=/bin/bash
export GOPRIVATE=github.com/anyproto
export PATH:=$(CURDIR)/deps:$(PATH)
export CGO_ENABLED:=1
BUILD_GOOS:=$(shell go env GOOS)
BUILD_GOARCH:=$(shell go env GOARCH)

ifeq ($(CGO_ENABLED), 0)
	TAGS:=-tags nographviz
else
	TAGS:=
endif

build:
	@$(eval FLAGS := $$(shell PATH=$(PATH) govvv -flags -pkg github.com/anyproto/any-sync/app))
	GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) go build -v $(TAGS) -o bin/any-sync-filenode -ldflags "$(FLAGS) -X github.com/anyproto/any-sync/app.AppName=any-sync-filenode" github.com/anyproto/any-sync-filenode/cmd

build-dev:
	@$(eval FLAGS := $$(shell PATH=$(PATH) govvv -flags -pkg github.com/anyproto/any-sync/app))
	go build -v $(TAGS) -o bin/any-sync-filenode.dev --tags dev -ldflags "$(FLAGS)" github.com/anyproto/any-sync-filenode/cmd

test:
	go test ./... --cover $(TAGS)

deps:
	go mod download
	go build -o deps google.golang.org/protobuf/cmd/protoc-gen-go
	go build -o deps github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto
	go build -o deps github.com/ahmetb/govvv
	go build -o deps go.uber.org/mock/mockgen

mocks:
	echo 'Generating mocks...'
	go generate ./...

PROTOC=protoc
PROTOC_GEN_GO=deps/protoc-gen-go
PROTOC_GEN_VTPROTO=deps/protoc-gen-go-vtproto

define generate_proto
	@echo "Generating Protobuf for directory: $(1)"
	$(PROTOC) \
    		--go_out=. --plugin protoc-gen-go="$(PROTOC_GEN_GO)" \
    		--go-vtproto_out=. --plugin protoc-gen-go-vtproto="$(PROTOC_GEN_VTPROTO)" \
    		--go-vtproto_opt=features=marshal+unmarshal+size \
    		--proto_path=$(1) $(wildcard $(1)/*.proto)
endef

proto:
	$(call generate_proto,index/indexproto/protos)
