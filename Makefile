.PHONY: build test deps build-dev
export GOPRIVATE=github.com/anytypeio
export PATH:=deps:$(PATH)
CGO_ENABLED=1

ifeq ($(CGO_ENABLED), 0)
	TAGS:=-tags nographviz
else
	TAGS:=
endif

build:
	@$(eval FLAGS := $$(shell PATH=$(PATH) govvv -flags -pkg github.com/anytypeio/any-sync/app))
	CGO_ENABLED=$(CGO_ENABLED) go build -v $(TAGS) -o bin/any-sync-filenode -ldflags "$(FLAGS)" github.com/anytypeio/any-sync-filenode/cmd

build-dev:
	@$(eval FLAGS := $$(shell PATH=$(PATH) govvv -flags -pkg github.com/anytypeio/any-sync/app))
	CGO_ENABLED=$(CGO_ENABLED) go build -v $(TAGS) -o bin/any-sync-filenode.dev --tags dev -ldflags "$(FLAGS)" github.com/anytypeio/any-sync-filenode/cmd

test:
	go test ./... --cover $(TAGS)

deps:
	go mod download
	go build -o deps github.com/ahmetb/govvv
	go build -o deps github.com/gogo/protobuf/protoc-gen-gogofaster

proto:
	protoc --gogofaster_out=:. index/redisindex/indexproto/protos/*.proto

