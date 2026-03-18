VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS = -ldflags "-X github.com/iskry/cask/internal/cli.Version=$(VERSION)"

.PHONY: build test lint vet install clean

build:
	go build $(LDFLAGS) -o cask ./cmd/cask

test:
	go test ./... -count=1

lint:
	golangci-lint run ./...

vet:
	go vet ./...

install:
	go install $(LDFLAGS) ./cmd/cask

clean:
	rm -f cask
