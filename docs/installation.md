# Installation

Synapse is distributed as a **single static binary** (no Node.js at runtime, no CGO). Choose one of the methods below.

## Requirements

- **macOS or Linux** for Homebrew and the `curl` install script
- **Windows amd64** for prebuilt releases (zip from GitHub Releases; no Homebrew install yet)
- **Go 1.22+** (see `go.mod`) only if you build from source

## Windows (prebuilt release)

1. Open **[Releases](https://github.com/calvinchengx/synapse/releases)** and download **`synapse_<version>_windows_amd64.zip`** for the version you want (or **Latest**).
2. Extract the zip. You should get an executable named like `synapse_v1.2.3_windows_amd64.exe`.
3. Run it from that folder, **or** rename it to `synapse.exe`, move it to a directory on your `PATH`, and open a new terminal.

Verify in **PowerShell** or **cmd** (use the real file name if you did not rename):

```text
.\synapse_v1.2.3_windows_amd64.exe version
```

After you rename to `synapse.exe` and put it on your `PATH`, you can run `synapse version`.

**Notes**

- Releases are **64-bit (amd64)** only; there is no Windows **arm64** build in the matrix yet.
- Homebrew and `scripts/install.sh` are Unix-oriented; use the zip above on Windows.

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
| Windows: `synapse` / `'synapse' is not recognized` | Add the folder containing `synapse.exe` to **Settings → System → About → Advanced system settings → Environment Variables → Path**, or invoke the full path to the `.exe`. |
| Old version after upgrade | Re-run `brew upgrade calvinchengx/tap/synapse` or rebuild from a fresh `git pull`. |
| Permission errors | Avoid `sudo` for `curl \| sh` unless you understand the script; prefer Homebrew or a user-writable prefix. |

Next: **[Getting started](getting-started.md)**.
