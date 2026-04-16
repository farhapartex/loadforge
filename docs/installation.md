# Installation

## Linux / macOS (recommended)

```bash
curl -fsSL https://github.com/farhapartex/loadforge/releases/latest/download/install.sh | sudo bash
```

This single command:

1. Downloads the latest `loadforge` and `loadforge-web` binaries to `/usr/local/bin/`
2. Creates `~/.loadforge/` with a default configuration
3. Registers and starts the web UI as a background service

Open [http://localhost:8090](http://localhost:8090) immediately — no extra commands needed.

Default credentials: **admin / admin**

---

## Manual download

Pre-built binaries are available on the [releases page](https://github.com/farhapartex/loadforge/releases) for:

| OS | Architecture |
|---|---|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |
| Windows | amd64 |

Download the binary for your platform, make it executable, and move it to a directory on your `PATH`.

---

## Verify installation

```bash
loadforge version
```
