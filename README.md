# LoadForge

> Developer-first HTTP load testing — run from the terminal or a browser-based UI.

LoadForge is an open-source HTTP load testing tool built for developers and QA engineers who want to stress-test their APIs without fighting complex configuration or expensive SaaS platforms. Write a simple YAML scenario, point it at your API, and LoadForge hammers it with concurrent workers while giving you live metrics in a clean terminal UI or a browser dashboard.

---

## Why LoadForge?

Most load testing tools are either too heavy (JMeter, Gatling) or too limited (ab, wrk). LoadForge sits in the middle:

- **Zero dependencies** — a single binary, no JVM, no Node, no Docker required
- **Two interfaces** — use the CLI for scripts and CI pipelines, or the web UI for interactive testing
- **YAML-first** — define your entire test scenario in a readable YAML file that lives in your repo
- **OpenAPI / Swagger support** — point at an API spec URL and LoadForge generates and runs the scenario for you
- **Multiple load profiles** — constant, ramp, step, and spike traffic patterns out of the box
- **Assertions & SLA thresholds** — define pass/fail criteria and catch regressions automatically
- **History tracking** — every run is persisted so you can compare results over time

---

## Features

| Feature | Description |
|---|---|
| **CLI load testing** | Run scenarios from YAML files with a live terminal UI |
| **Web UI** | Browser-based dashboard to trigger, monitor, and review tests |
| **OpenAPI / Swagger import** | Auto-generate scenarios from OpenAPI 3.x and Swagger 2.0 specs |
| **HAR file support** | Convert or replay browser HAR recordings as load tests |
| **Load profiles** | `constant`, `ramp`, `step`, `spike` — choose how traffic behaves |
| **Assertions** | Define SLA thresholds (p95 latency, error rate, RPS) with pass/fail results |
| **Auth support** | Basic auth, Bearer tokens, and custom header auth per request step |
| **Variable injection** | Pass variables via `--var` flags or a `.env` file |
| **History** | Persistent run history with per-request breakdown |
| **Password protection** | Secure the web UI with bcrypt-hashed credentials |
| **Multi-platform** | Linux, macOS, and Windows — amd64 and arm64 |

---

## Installation

### Linux / macOS (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/farhapartex/loadforge/main/scripts/install.sh | sudo bash
```

This installs both `loadforge` (CLI) and `loadforge-web` (web UI server) to `/usr/local/bin/` and creates `~/.loadforge/` with a default configuration.

### Manual download

Download the binaries for your platform from the [latest release](https://github.com/farhapartex/loadforge/releases/latest) and place them in your `$PATH`.

---

## Uninstallation

```bash
sudo loadforge --uninstall
```

This removes both binaries from `/usr/local/bin/` and deletes the `~/.loadforge/` data directory.

---

## Quick Start

### Web UI

```bash
loadforge-web
```

Open `http://localhost:8080` in your browser. Default credentials: `admin` / `admin`.

Paste an OpenAPI or Swagger spec URL, set your load parameters, and click **Start Test**. Live logs stream to the dashboard. Past runs appear under **History**.

### CLI

```bash
# Run a scenario file
loadforge run my-scenario.yaml

# Override workers and duration at runtime
loadforge run my-scenario.yaml --workers 50 --duration 2m

# Inject variables
loadforge run my-scenario.yaml --var base_url=https://api.example.com

# Load variables from a .env file
loadforge run my-scenario.yaml --env-file .env

# Disable the terminal UI (plain output, useful in CI)
loadforge run my-scenario.yaml --no-ui

# Validate a scenario without running it
loadforge validate my-scenario.yaml

# Check installed version
loadforge version
```

---

## Scenario File

LoadForge scenarios are plain YAML files you can commit alongside your code.

```yaml
name: user-api-test
base_url: https://api.example.com

load:
  profile: ramp
  duration: 2m
  ramp_up:
    start_workers: 5
    end_workers: 50
    duration: 30s

scenarios:
  - name: create-and-fetch-user
    weight: 1
    steps:
      - name: create user
        method: POST
        url: /users
        headers:
          Content-Type: application/json
        body:
          json:
            name: "Test User"
            email: "test@example.com"
        auth:
          bearer: "your-token-here"

      - name: get users
        method: GET
        url: /users
        think: 500ms

assertions:
  - metric: p95_latency
    operator: less_than
    value: 500
    enabled: true
  - metric: error_rate
    operator: less_than
    value: 1
    enabled: true
```

---

## Load Profiles

| Profile | Use case |
|---|---|
| `constant` | Fixed workers for the full duration — steady baseline benchmark |
| `ramp` | Gradually increase workers — find the degradation point without shock |
| `step` | Add workers in increments — observe how latency changes as load grows |
| `spike` | Base workers with periodic bursts — test recovery after a traffic surge |

---

## Assertions

Assertions let you define SLA thresholds that automatically mark a test as passed or failed.

**Available metrics:** `p50_latency`, `p90_latency`, `p95_latency`, `p99_latency`, `avg_latency`, `max_latency`, `rps`, `error_rate`, `success_rate`, `total_requests`, `total_errors`

**Available operators:** `less_than`, `less_than_or_equal`, `greater_than`, `greater_than_or_equal`, `equal`

---

## Web UI Configuration

The web UI reads its configuration from `~/.loadforge/web.yml`. You can edit this file to change the address, credentials, or session settings.

```yaml
addr: ":8080"
username: "admin"
password: "admin"
session_ttl: "24h"
log_file: "/Users/you/.loadforge/load_forge.logs"
history_file: "/Users/you/.loadforge/load_forge_history.json"
```

To change your password securely, use the **Settings** page in the web UI — the new password is stored as a bcrypt hash.

---

## Building from Source

Requires Go 1.26+.

```bash
git clone https://github.com/farhapartex/loadforge.git
cd loadforge

# Build CLI
make build          # → ./bin/loadforge

# Build web UI server
make build-web      # → ./bin/loadforge-web

# Run directly without building
make run ARGS="run my-scenario.yaml"
make run-web
```

---

## Feedback & Support

Have a question, found a bug, or want to suggest a feature?

- **Email:** [hasan08sust@gmail.com](mailto:hasan08sust@gmail.com)
- **GitHub Issues:** [github.com/farhapartex/loadforge/issues](https://github.com/farhapartex/loadforge/issues)

Contributions are welcome. Open a pull request or start a discussion in Issues.

---

## License

LoadForge is released under the [MIT License](LICENSE).
