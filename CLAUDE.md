# Synapse - Development Guide

## Build & Test

```bash
make build          # Build binary with embedded frontend stub
make test           # Run all tests with race detector
make test-v         # Verbose tests
make coverage       # Generate coverage report (coverage.html)
make lint           # Run golangci-lint
make frontend       # Build Svelte frontend (requires npm)
make dev            # Run Go server in development mode
make frontend-dev   # Run Vite dev server (in separate terminal)
```

## Architecture

- **Pure Go, no CGO.** Uses `modernc.org/sqlite` for SQLite access.
- **YAML only** for configuration and manifests. No TOML.
- **Gin** for HTTP server (`internal/server/`).
- **Cobra** for CLI (`cmd/synapse/`).
- Frontend is Svelte 5 embedded via `//go:embed` in `internal/web/embed.go`.

## Project Layout

- `cmd/synapse/` — CLI entry point
- `internal/config/` — YAML config loading with env overrides
- `internal/rules/` — Rule engine, scanner, semantic router, installer
- `internal/integration/` — Tool manifests, discovery, data adapters, events
- `internal/server/` — Gin HTTP server, REST API, middleware
- `internal/deploy/` — Deploy rules to .claude/, .cursor/, .codex/
- `internal/llm/` — LLM client (LiteLLM Docker proxy, Anthropic, OpenAI)
- `internal/errors/` — Typed errors, circuit breaker
- `internal/db/` — Synapse metadata SQLite
- `internal/web/` — Embedded frontend assets
- `config/` — Bundled rules, skills, commands, agents, contexts
- `frontend/` — Svelte 5 SPA source
- `tests/` — E2E tests and test utilities

## Testing Conventions

- Every package gets a `_test.go` file.
- Table-driven tests for anything with multiple cases.
- Use `t.TempDir()` for filesystem tests.
- Use `httptest.NewServer` for HTTP tests.
- Use in-memory SQLite for database tests.
- Race detector always on: `go test -race`.
- Target 80%+ coverage per package.

## LiteLLM Safety

**CRITICAL:** Only use the official LiteLLM Docker image (`ghcr.io/berriai/litellm`).
Never `pip install litellm`. PyPI versions 1.82.7 and 1.82.8 were compromised.
The Docker image was NOT affected. See docs/PLAN.md for details.
