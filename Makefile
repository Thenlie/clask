PKGS := ./cmd/... ./internal/...
BIN  := bin/clask
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-s -w -X github.com/Thenlie/clask/internal/cli.version=$(VERSION)"

.PHONY: build test vet fmt install clean

build:
	go build $(LDFLAGS) -o $(BIN) ./cmd/clask

test: vet
	go test $(PKGS)

vet:
	go vet $(PKGS)

fmt:
	gofmt -l -w cmd internal

install:
	go install $(LDFLAGS) ./cmd/clask

clean:
	rm -rf bin dist
