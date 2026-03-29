# Synapse

> The psychic link between AI coding tools.

Synapse is a Go CLI + embedded web UI that manages rules, skills, and agents for AI coding assistants (Claude Code, Cursor, Codex), and connects independent tools like [RTK](https://github.com/rtk-ai/rtk) and [AgentsView](https://github.com/wesm/agentsview) into a unified workflow.

Built in Go — single binary, no Node.js runtime, no CGO.

## Documentation

| | |
|--|--|
| **[Installation](docs/installation.md)** | Homebrew, install script, build from source |
| **[Getting started](docs/getting-started.md)** | First run: `init`, daily commands, rule files |
| **[Configuration](docs/configuration.md)** | `~/.synapse/config.yaml`, env vars, LiteLLM |
| **[CLI reference](docs/cli-reference.md)** | All commands and flags |
| **[Doc site (local)](docs/index.md)** | Build searchable HTML with `make docs` (requires Python + MkDocs) |

**GitHub Pages:** After you [enable Pages](#github-pages-for-documentation), pushes to `main` that touch `docs/`, `mkdocs.yml`, or `pyproject.toml` / `uv.lock` build and deploy the Material site to `https://<your-username>.github.io/synapse/` (forks: adjust `site_url` in `mkdocs.yml`).

Internal design notes: [docs/PLAN.md](docs/PLAN.md).

---

## Contents

- [What it does](#what-it-does)
- [Install](#install)
- [Quick start](#quick-start)
- [Commands (summary)](#commands-summary)
- [Configuration (summary)](#configuration-summary)
- [LiteLLM](#litellm)
- [Rule format](#rule-format)
- [Integrations](#integrations)
- [Plugin system](#plugin-system)
- [Development](#development)
- [Architecture](#architecture)
- [License](#license)

---

## What it does

- **Rule management** — maintain a library of rules, skills, agents, and contexts as Markdown files. Deploy them to `.claude/`, `.cursor/`, or `.codex/AGENTS.md` with one command.
- **Semantic routing** — per-prompt, activates only the rules relevant to what you are working on. Uses an LLM (Tier 1), keyword matching (Tier 2), and integration signals (Tier 3) to decide.
- **Integrations** — reads RTK token savings and AgentsView session data from their SQLite databases and surfaces aggregated reports.
- **Web UI** — embedded Svelte SPA served from the binary for browsing rules and integration status.
- **Hooks** — `synapse hook` replaces the Node.js `semantic-router.cjs` hook. Drop it in `.claude/settings.json` and it runs on every prompt with zero external runtime.

---

## Install

Full guide: **[docs/installation.md](docs/installation.md)**.

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

Step-by-step: **[docs/getting-started.md](docs/getting-started.md)**.

```bash
synapse init
synapse list
synapse test "add authentication middleware"
synapse serve
synapse integrations
synapse doctor
```

Use `synapse serve` and open the printed URL for the API and UI (`browse` may not be implemented yet).

---

## Commands (summary)

Full table: **[docs/cli-reference.md](docs/cli-reference.md)**.

| Command | Description |
|---------|-------------|
| `synapse init` | Initialise rules in the current project |
| `synapse list` | List all installed rules by category |
| `synapse search <query>` | Search rules by keyword |
| `synapse test <prompt>` | Preview which rules the router would activate |
| `synapse add <url>` | Add a rule source (planned) |
| `synapse update` | Update rules from all sources |
| `synapse serve` | Start the HTTP API server |
| `synapse browse` | Open the web UI (planned) |
| `synapse integrations` | Show detected third-party integrations |
| `synapse doctor` | Run diagnostics |
| `synapse hook` | Claude Code hook (stdin JSON); hidden in `--help` |
| `synapse version` | Print version |

---

## Configuration (summary)

Reference: **[docs/configuration.md](docs/configuration.md)**.

Synapse reads `~/.synapse/config.yaml`, falling back to built-in defaults. Every field listed in the docs can be overridden with an environment variable where noted.

```yaml
sources:
  - path: ./config
targets:
  - claude
  - cursor
  - codex
router:
  always_active:
    - essential.md
    - security.md
  max_active_rules: 10
llm:
  provider: litellm
  base_url: http://localhost:4000
  model: claude-3-haiku
  timeout: 10s
  max_retries: 2
server:
  host: 127.0.0.1
  port: 8080
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

Synapse can use LiteLLM as a proxy. **Only use the official Docker image** — the PyPI package was compromised in March 2026 (versions 1.82.7 and 1.82.8). The Docker image was not affected.

```bash
docker run -p 4000:4000 ghcr.io/berriai/litellm \
  --model claude-3-haiku \
  --api_key $ANTHROPIC_API_KEY
```

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
plugin.yaml
rules/
skills/
agents/
contexts/
hooks/
integrations/
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

Documentation site ([uv](https://docs.astral.sh/uv/) + `pyproject.toml`):

```bash
uv sync --group docs
make docs-serve     # live reload at http://127.0.0.1:8000
make docs           # static site in ./site
```

### GitHub Pages (documentation)

Do this **once** in the GitHub repo:

1. **Settings → Pages** (left sidebar).
2. Under **Build and deployment**, set **Source** to **GitHub Actions** (not “Deploy from a branch”).
3. Merge or push `.github/workflows/docs.yml` to `main`. The **Deploy documentation** workflow runs; open **Actions** to confirm it succeeded.
4. After the first deploy, **Settings → Pages** shows the public URL (for a project site: `https://<owner>.github.io/<repo>/`).

If the URL differs (fork or custom domain), set `site_url` in `mkdocs.yml` to match so search and canonical links are correct.

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
├── docs/                 user guides + planning (MkDocs sources)
├── pyproject.toml        uv dependency groups for MkDocs Material
└── uv.lock               locked doc-tooling versions (CI: uv sync --frozen)
```

---

## License

MIT
