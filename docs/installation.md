# Installation

Synapse is distributed as a **single static binary** (no Node.js at runtime, no CGO). Choose one of the methods below.

## Requirements

- **macOS or Linux** for prebuilt releases and Homebrew
- **Go 1.22+** (see `go.mod`) only if you build from source

## Homebrew (macOS / Linux)

```bash
brew install calvinchengx/tap/synapse
```

Verify:

```bash
synapse version
```

## Install script (curl)

```bash
curl -fsSL https://raw.githubusercontent.com/calvinchengx/synapse/main/scripts/install.sh | sh
```

Review the script on GitHub before piping to `sh` if you prefer.

## Build from source

```bash
git clone https://github.com/calvinchengx/synapse.git
cd synapse
make build
```

This produces `./synapse` in the repository root. Optionally install it somewhere on your `PATH`:

```bash
cp synapse /usr/local/bin/   # example
```

### Frontend in the binary

- `make build` embeds a **stub** HTML shell if the Svelte app has not been built yet.
- For a release build with the full SPA embedded, run `make build-release` (requires Node.js and npm).

## Troubleshooting

| Issue | What to try |
|--------|-------------|
| `command not found: synapse` | Ensure the install location is on your `PATH`, or use `./synapse` from the build directory. |
| Old version after upgrade | Re-run `brew upgrade calvinchengx/tap/synapse` or rebuild from a fresh `git pull`. |
| Permission errors | Avoid `sudo` for `curl \| sh` unless you understand the script; prefer Homebrew or a user-writable prefix. |

Next: **[Getting started](getting-started.md)**.
