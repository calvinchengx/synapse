# Synapse

> The psychic link between AI coding tools.

Synapse is a Go CLI + embedded web UI that manages rules, skills, and agents for AI coding assistants (Claude Code, Cursor, Codex), and connects independent tools like [RTK](https://github.com/rtk-ai/rtk) and [AgentsView](https://github.com/wesm/agentsview) into a unified workflow.

It is a full rewrite of [ai-nexus](https://github.com/JSK9999/ai-nexus) in Go — single binary, no Node.js runtime, no CGO.

---

## What it does

- **Rule management** — maintain a library of rules, skills, agents, and contexts as Markdown files. Deploy them to `.claude/`, `.cursor/`, or `.codex/AGENTS.md` with one command.
- **Semantic routing** — per-prompt, activates only the rules relevant to what you are working on. Uses an LLM (Tier 1), keyword matching (Tier 2), and integration signals (Tier 3) to decide.
- **Integrations** — reads RTK token savings and AgentsView session data from their SQLite databases and surfaces aggregated reports.
- **Web UI** — embedded Svelte SPA served from the binary for browsing rules and integration status.
- **Hooks** — `synapse hook` replaces the Node.js `semantic-router.cjs` hook. Drop it in `.claude/settings.json` and it runs on every prompt with zero external runtime.

---

## Install

### Homebrew (macOS / Linux)

```bash
brew install calvinchengx/tap/synapse
```

### curl

```bash
curl -fsSL https://raw.githubusercontent.com/calvinchengx/synapse/main/scripts/install.sh | sh
```

### From source

```bash
git clone https://github.com/calvinchengx/synapse
cd synapse
make build          # produces ./synapse binary
```

---

## Quick start

```bash
# Initialise in your project (deploys rules to .claude/, .cursor/, .codex/)
synapse init

# List installed rules
synapse list

# Test the semantic router against a prompt
synapse test "add authentication middleware"

# Open the web UI
synapse browse

# Check integration status
synapse integrations

# Run diagnostics
synapse doctor
```

---

## Commands

| Command | Description |
|---------|-------------|
| `synapse init` | Initialise rules in the current project |
| `synapse list` | List all installed rules by category |
| `synapse search <query>` | Search rules by keyword |
| `synapse test <prompt>` | Preview which rules the router would activate |
| `synapse add <url>` | Add a rule source (git URL or local path) |
| `synapse update` | Update rules from all sources |
| `synapse serve` | Start the HTTP API server |
| `synapse browse` | Start the server and open the web UI |
| `synapse integrations` | Show detected third-party integrations |
| `synapse doctor` | Run diagnostics |
| `synapse hook` | Claude Code hook executor (used in settings.json) |
| `synapse version` | Print version |

---

## Configuration

Synapse reads `~/.synapse/config.yaml`, falling back to built-in defaults. Every field can be overridden with an environment variable.

```yaml
sources:
  - path: ./config          # local rule directory
  - url: https://github.com/org/rules  # external rule repo

targets:
  - claude                  # deploy to .claude/
  - cursor                  # deploy to .cursor/
  - codex                   # deploy to .codex/AGENTS.md

router:
  always_active:
    - essential.md
    - security.md
  max_active_rules: 10

llm:
  provider: litellm         # litellm | anthropic | openai
  base_url: http://localhost:4000
  model: claude-3-haiku
  timeout: 10s
  max_retries: 2

server:
  host: 127.0.0.1
  port: 8090
```

| Environment variable | Overrides |
|---------------------|-----------|
| `SYNAPSE_HOST` | `server.host` |
| `SYNAPSE_PORT` | `server.port` |
| `SYNAPSE_LLM_PROVIDER` | `llm.provider` |
| `SYNAPSE_LLM_BASE_URL` | `llm.base_url` |
| `SYNAPSE_LLM_MODEL` | `llm.model` |

---

## LiteLLM

Synapse uses LiteLLM as a proxy to route LLM calls. **Only use the official Docker image** — the PyPI package was compromised in March 2026 (versions 1.82.7 and 1.82.8). The Docker image was not affected.

```bash
docker run -p 4000:4000 ghcr.io/berriai/litellm \
  --model claude-3-haiku \
  --api_key $ANTHROPIC_API_KEY
```

See [docs/PLAN.md](docs/PLAN.md) for the full architecture and implementation notes.

---

## Rule format

Rules, skills, agents, and contexts are plain Markdown with optional YAML frontmatter.

```markdown
---
description: Security guidelines for web APIs
keywords: [security, auth, jwt, owasp]
tools: Read, Grep, Bash
alwaysApply: false
---

# Security

Always validate and sanitise user input at system boundaries...
```

Place files under `config/rules/`, `config/skills/`, `config/agents/`, or `config/contexts/`. Run `synapse init` to deploy.

---

## Integrations

Synapse discovers and reads from third-party tools automatically.

| Tool | What synapse reads |
|------|--------------------|
| **RTK** | Token savings history from `tracking.db` — aggregated by command and time period |
| **AgentsView** | Session analytics from `agentsview.db` — tool usage distribution, session count |

Run `synapse integrations` to see detection status. No configuration needed if the tools are installed in standard locations.

---

## Plugin system

External rule packages can be added as plugins:

```bash
synapse plugin add https://github.com/org/synapse-plugin-mystack
```

A plugin is a git repository containing:

```
plugin.yaml          # plugin manifest
rules/               # rules and guidelines
skills/              # language/framework skills
agents/              # sub-agent definitions
contexts/            # session contexts
hooks/               # optional shell hook scripts
integrations/        # optional tool manifests
```

---

## Development

```bash
make build          # build binary (embeds frontend stub)
make test           # run all tests with race detector
make coverage       # generate coverage.html
make lint           # run golangci-lint
make frontend       # build Svelte frontend
make dev            # run Go server in dev mode (port 8090)
make frontend-dev   # run Vite dev server (port 5173, in separate terminal)
```

Tests require no external services — SQLite uses in-memory databases, HTTP uses `httptest`, LLM calls use mock interfaces.

---

## Architecture

```
synapse/
├── cmd/synapse/          CLI entry point (cobra)
├── internal/
│   ├── config/           YAML config with env overrides
│   ├── rules/            engine, scanner, router, installer, registry
│   ├── deploy/           write to .claude/, .cursor/, .codex/
│   ├── integration/      manifest loader, discovery, SQLite readers, events
│   ├── llm/              LiteLLM / Anthropic / OpenAI client
│   ├── server/           Gin HTTP server + REST API
│   ├── errors/           typed errors, circuit breaker
│   └── web/              go:embed for Svelte dist
├── config/               built-in rules, skills, agents, contexts
├── frontend/             Svelte 5 SPA source
└── docs/                 architecture and planning docs
```

---

## License

MIT
