.PHONY: build install clean test release release-snapshot

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -s -w \
	-X github.com/meshbrow-dev/meshbrow-cli/cmd.Version=$(VERSION) \
	-X github.com/meshbrow-dev/meshbrow-cli/cmd.Commit=$(COMMIT) \
	-X github.com/meshbrow-dev/meshbrow-cli/cmd.BuildDate=$(DATE)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/meshbrow .

install: build
	cp bin/meshbrow /usr/local/bin/meshbrow
	@echo "✓ meshbrow installed to /usr/local/bin/meshbrow"

test:
	go test ./... -count=1

clean:
	rm -rf bin/ dist/

release:
	goreleaser release --clean

release-snapshot:
	goreleaser release --snapshot --clean
