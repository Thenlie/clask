PKGS := ./cmd/... ./internal/...
BIN  := bin/clask
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-s -w -X github.com/Thenlie/clask/internal/cli.version=$(VERSION)"

.PHONY: build build-go test vet fmt install clean web web-check web-dev

# Full binary: compiles the web UI, then embeds it.
build: web build-go

# Go only; serves the "UI not built" page unless dist/ was built earlier.
build-go:
	go build $(LDFLAGS) -o $(BIN) ./cmd/clask

test: vet
	go test $(PKGS)

vet:
	go vet $(PKGS)

fmt:
	gofmt -l -w cmd internal

# `make web install` for a binary with the UI.
install:
	go install $(LDFLAGS) ./cmd/clask

web:
	pnpm --dir web install --frozen-lockfile
	pnpm --dir web build

web-check:
	pnpm --dir web typecheck

# Vite dev server with hot reload; run `clask serve --no-open` alongside.
web-dev:
	pnpm --dir web dev

clean:
	rm -rf bin dist internal/server/dist/assets internal/server/dist/index.html internal/server/dist/favicon.svg
