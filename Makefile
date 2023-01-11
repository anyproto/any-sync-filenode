.PHONY: build test deps
export GOPRIVATE=github.com/anytypeio
export PATH:=deps:$(PATH)

build:
	@$(eval FLAGS := $$(shell PATH=$(PATH) govvv -flags -pkg github.com/anytypeio/any-sync/app))
	go build -v -o bin/any-sync-filenode -ldflags "$(FLAGS)" github.com/anytypeio/any-sync-filenode/cmd

test:
	go test ./... --cover

deps:
	go mod download
	go build -o deps github.com/ahmetb/govvv
