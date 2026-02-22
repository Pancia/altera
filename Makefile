.PHONY: install build test lint clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS  = -X github.com/anthropics/altera/internal/cli.Version=$(VERSION) \
           -X github.com/anthropics/altera/internal/cli.Commit=$(COMMIT)

install: build
	cp bin/alt ~/.local/bin/alt
	codesign -s - ~/.local/bin/alt

build:
	go build -ldflags "$(LDFLAGS)" -o bin/alt ./cmd/alt

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
