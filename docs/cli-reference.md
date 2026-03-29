# CLI reference

All commands use the `synapse` binary. Global help:

```bash
synapse --help
synapse <command> --help
```

## Commands

### `synapse init`

Initialise Synapse in the **current working directory**: metadata, load rules from configured sources, deploy to configured targets (Claude, Cursor, Codex).

| Flag | Description |
|------|-------------|
| `-i`, `--interactive` | Reserved for interactive setup (wizard) |

---

### `synapse update`

Refresh rules from sources, update the install, and redeploy to targets.

---

### `synapse list`

Print all installed rules grouped by **category**, with truncated descriptions.

---

### `synapse search <keyword>`

Search **local** rules whose metadata matches the keyword.

---

### `synapse test <prompt>`

Run the **semantic router** against a prompt and print which rules would be activated (tier, method, and list). Uses the LLM from config when `llm.provider` is set; otherwise falls back to non-LLM tiers.

---

### `synapse doctor`

Print diagnostics: config file presence, rule source paths, data directory, LLM settings.

---

### `synapse serve`

Start the **HTTP server** (REST API + embedded SPA). Does not open a browser.

| Flag | Default | Description |
|------|---------|-------------|
| `--host` | `127.0.0.1` | Bind address |
| `--port` | `8090` | Listen port |

---

### `synapse integrations`

List configured integrations and whether tools (binaries, APIs) appear available.

---

### `synapse version`

Print version, git commit, and build date embedded at compile time.

---

### `synapse add <url>` / `synapse remove <source>` / `synapse get <file>`

Planned or stubbed commands for adding/removing sources and fetching single files. Behaviour may print “not yet implemented” until completed.

---

### `synapse browse`

Planned: start the server and open the web UI. May be unimplemented; use `synapse serve` and open the URL manually.

---

### `synapse uninstall`

Remove Synapse-managed files from the **current project**.

---

### `synapse hook` (advanced)

**Hidden** from default `--help`. Intended as a **Claude Code hook**: reads JSON from stdin, runs the router, and moves rule files between active/inactive directories under `.claude/`. Configure in Claude Code `settings.json` as the hook command pointing at the `synapse` binary with the `hook` subcommand.

---

## Configuration file

User-wide defaults: `~/.synapse/config.yaml`. See **[Configuration](configuration.md)**.
