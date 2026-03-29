# Getting started

This guide walks through your first use of Synapse in a project: initialise rules, inspect the router, run the web server, and understand where files live.

## Before you begin

- [Install Synapse](installation.md)
- Optional: configure [LiteLLM or another LLM](configuration.md#llm-and-litellm) for Tier-1 semantic routing (`synapse test`). Keyword routing works without an LLM.

## Initialise a project

From your repository root:

```bash
synapse init
```

This will:

1. Create Synapse metadata under `~/.synapse/` (and project-specific state as implemented).
2. Load rules from configured **sources** (by default `./config` relative to the current directory).
3. **Deploy** rules to assistant targets (for example `.claude/`, `.cursor/`, `.codex/`) according to `~/.synapse/config.yaml`.

If a source path is missing, `synapse doctor` will report it.

## Everyday commands

```bash
# List rules grouped by category
synapse list

# Search local rules by keyword
synapse search jwt

# See which rules the router would activate for a prompt
synapse test "add OAuth2 to the API"

# Check config paths, sources, and LLM settings
synapse doctor

# HTTP API + embedded UI (default http://127.0.0.1:8090)
synapse serve

# Third-party tools (RTK, AgentsView, etc.)
synapse integrations
```

`browse` (open browser to the UI) may be unavailable until implemented; use `synapse serve` and open the URL manually.

## Update deployed rules

After you change rule files or pull updates:

```bash
synapse update
```

This refreshes installed copies and redeploys to your configured targets.

## Remove Synapse from a project

```bash
synapse uninstall
```

Removes Synapse-managed files from the **current** project directory (does not delete your global `~/.synapse` config unless documented otherwise).

## Rule files

Rules, skills, agents, and contexts are **Markdown** with optional **YAML frontmatter**.

```markdown
---
description: Short summary for lists and routing
keywords: [api, security, jwt]
tools: Read, Grep, Bash
alwaysApply: false
---

# Title

Your guidance for the assistant...
```

Typical layout in a rule package:

- `config/rules/` — coding standards, workflows
- `config/skills/` — language or stack skills
- `config/agents/` — sub-agent definitions
- `config/contexts/` — session contexts

Run `synapse init` after adding or editing files so they are installed and deployed.

## Next steps

- Tune **[Configuration](configuration.md)** (sources, targets, router, server).
- See **[CLI reference](cli-reference.md)** for every command and flag.
