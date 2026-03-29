# Synapse documentation

Welcome to the documentation for **Synapse** — a single-binary CLI and embedded web UI that manages rules, skills, and agents for AI coding assistants (Claude Code, Cursor, Codex) and connects tools like RTK and AgentsView into one workflow.

## Guides

- **[Installation](installation.md)** — Homebrew, install script, and building from source
- **[Getting started](getting-started.md)** — First project setup, daily commands, rule files
- **[Configuration](configuration.md)** — `~/.synapse/config.yaml`, environment variables, LiteLLM
- **[CLI reference](cli-reference.md)** — All commands, flags, and behavior notes

## Internal

- **[Architecture & plan](PLAN.md)** — Design notes and roadmap (for contributors)

## Source

Documentation Markdown files live in the `docs/` directory of the [Synapse repository](https://github.com/calvinchengx/synapse). Browse them on GitHub or build a local searchable site with [MkDocs](https://www.mkdocs.org/) using [uv](https://docs.astral.sh/uv/) and `pyproject.toml` (see the repository `Makefile`).
