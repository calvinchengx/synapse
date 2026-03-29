VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X main.version=$(VERSION) \
           -X main.commit=$(COMMIT) \
           -X main.buildDate=$(BUILD_DATE)
LDFLAGS_RELEASE := $(LDFLAGS) -s -w

# Ensure go:embed has content even before first frontend build
ensure-embed-dir:
	@mkdir -p internal/web/dist
	@test -f internal/web/dist/stub.html || echo '<!doctype html><title>build required</title>' > internal/web/dist/stub.html

# Build frontend and copy to embed directory
frontend:
	cd frontend && npm install && npm run build
	rm -rf internal/web/dist
	cp -r frontend/dist internal/web/dist

# Development: Go server with stub frontend
dev: ensure-embed-dir
	go run ./cmd/synapse serve

# Development: Vite dev server (run alongside `make dev`)
frontend-dev:
	cd frontend && npm run dev

# Debug build with embedded frontend
build: ensure-embed-dir
	go build -ldflags="$(LDFLAGS)" -o synapse ./cmd/synapse

# Release build: stripped, trimmed
build-release: frontend
	go build -ldflags="$(LDFLAGS_RELEASE)" -trimpath -o synapse ./cmd/synapse

# Tests (use stub frontend, don't require npm)
test: ensure-embed-dir
	go test ./... -race -count=1

# Verbose tests
test-v: ensure-embed-dir
	go test ./... -v -race -count=1

# Coverage report
coverage: ensure-embed-dir
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

# User documentation (MkDocs Material). Requires: uv (https://docs.astral.sh/uv/)
docs:
	uv run --group docs mkdocs build

docs-serve:
	uv run --group docs mkdocs serve

clean:
	rm -f synapse coverage.out coverage.html
	rm -rf internal/web/dist

help:
	@echo "Synapse — available targets:"
	@echo "  make build           Debug binary (stub frontend if no npm build yet)"
	@echo "  make build-release   Release binary: npm build + stripped Go binary"
	@echo "  make dev             Run synapse serve with stub frontend"
	@echo "  make frontend        npm install + build; copy dist into internal/web"
	@echo "  make frontend-dev    Vite dev server (use alongside make dev)"
	@echo "  make test            Run tests (-race)"
	@echo "  make test-v          Verbose tests"
	@echo "  make coverage        Tests + coverage.html"
	@echo "  make lint            golangci-lint"
	@echo "  make docs            MkDocs build (needs uv)"
	@echo "  make docs-serve      MkDocs live server"
	@echo "  make clean           Remove synapse binary, coverage artifacts, web dist"
	@echo "  make ensure-embed-dir  Ensure internal/web/dist exists for embed"

.PHONY: help ensure-embed-dir frontend dev frontend-dev build build-release test test-v coverage lint docs docs-serve clean
