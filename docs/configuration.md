# Configuration

Synapse reads **`~/.synapse/config.yaml`**. If the file is missing, built-in defaults are used. Certain fields can be overridden with **environment variables** (applied after loading YAML).

## Example `config.yaml`

```yaml
sources:
  - path: ./config
  # - url: https://github.com/org/rules
  #   branch: main

targets:
  - claude    # .claude/
  - cursor    # .cursor/
  - codex     # .codex/AGENTS.md

router:
  always_active:
    - essential.md
    - security.md
  max_active_rules: 10

llm:
  provider: litellm          # litellm | anthropic | openai
  base_url: http://localhost:4000
  api_key_env: ANTHROPIC_API_KEY
  model: claude-3-haiku
  timeout: 10s
  max_retries: 2

server:
  host: 127.0.0.1
  port: 8080
  auth_token: ""
  allowed_origins: []
  allowed_hosts: []

integrations:
  rtk:
    enabled: true
  agentsview:
    enabled: true
    api_url: http://localhost:58080
```

## Environment variables

These override the corresponding fields after the file is loaded:

| Variable | Overrides |
|----------|-----------|
| `SYNAPSE_HOST` | `server.host` |
| `SYNAPSE_PORT` | `server.port` |
| `SYNAPSE_LLM_PROVIDER` | `llm.provider` |
| `SYNAPSE_LLM_BASE_URL` | `llm.base_url` |
| `SYNAPSE_LLM_MODEL` | `llm.model` |

## LLM and LiteLLM

Semantic routing can call an LLM (Tier 1). A common setup is **LiteLLM** as a local proxy.

**Security:** Use only the **official LiteLLM Docker image** (`ghcr.io/berriai/litellm`). Do not `pip install` LiteLLM from PyPI for this use case; specific PyPI versions were compromised in 2026 — the Docker image was not affected.

Example:

```bash
docker run -p 4000:4000 ghcr.io/berriai/litellm \
  --model claude-3-haiku \
  --api_key "$ANTHROPIC_API_KEY"
```

Point `llm.base_url` at your proxy (for example `http://localhost:4000`) and set `llm.provider` / `llm.model` to match your deployment.

## `synapse serve` flags

The `serve` command accepts **command-line flags** that control the listening address regardless of `server.host` / `server.port` in YAML:

- `--host` (default `127.0.0.1`)
- `--port` (default `8090`)

Use `synapse serve --help` for details.

## Integrations

Integration manifests can live in:

- `./config/integrations`
- `~/.synapse/integrations`
- `.synapse/integrations` (project)

See **[Getting started](getting-started.md)** and `synapse integrations` for discovery status.
