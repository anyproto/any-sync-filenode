.PHONY: proto build test
export GOPRIVATE=github.com/anytypeio

build:
	@$(eval FLAGS := $$(shell govvv -flags -pkg github.com/anytypeio/go-anytype-infrastructure-experiments/filenode))
	go build -v -o bin/any-sync-filenode -ldflags "$(FLAGS)" github.com/anytypeio/any-sync-filenode/cmd

test:
	go test ./... --cover
