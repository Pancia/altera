.PHONY: build test lint clean

build:
	go build -o bin/alt ./cmd/alt

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
