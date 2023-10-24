.PHONY: build test deps build-dev
SHELL=/bin/bash
export GOPRIVATE=github.com/anyproto
export PATH:=deps:$(PATH)
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
	go build -o deps github.com/ahmetb/govvv
	go build -o deps github.com/gogo/protobuf/protoc-gen-gogofaster

proto:
	protoc --gogofaster_out=:. index/indexproto/protos/*.proto
