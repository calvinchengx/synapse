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

clean:
	rm -f synapse coverage.out coverage.html
	rm -rf internal/web/dist

.PHONY: ensure-embed-dir frontend dev frontend-dev build build-release test test-v coverage lint clean
