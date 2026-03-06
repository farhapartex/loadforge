# loadforge

A developer-first HTTP load testing tool. Single binary, zero external dependencies to install, config-driven via YAML so your tests are repeatable and version-controllable.

---

## What is loadforge?

loadforge lets you define HTTP load tests as YAML scenarios and run them with different traffic profiles — constant load, gradual ramp-up, step increases, or sudden spikes. It ships with a real-time terminal dashboard that shows live metrics as the test runs. It is designed for developers and platform engineers who want to stress-test APIs without reaching for a heavy GUI tool.

---

## Project Status

This project is in active development. The core engine and all four load profiles are functional. The terminal UI, reporting, and HAR file features are not yet implemented.

| Feature | Status |
|---|---|
| YAML config parsing & validation | Done |
| HTTP engine (GET, POST, PUT, PATCH, DELETE, HEAD) | Done |
| Request body — raw, JSON, form-encoded | Done |
| Auth — Basic, Bearer, custom header | Done |
| HTTP/2 support | Done |
| TLS skip verify | Done |
| Per-request timeout & redirect config | Done |
| Load profile — constant | Done |
| Load profile — ramp | Done |
| Load profile — step | Done |
| Load profile — spike | Done |
| Metrics — RPS, latencies, status codes, error rate | Done |
| Latency percentiles — p50, p90, p95, p99 | Done |
| `run` CLI command | Done |
| `validate` CLI command | Done |
| `version` CLI command | Done |
| Real-time terminal UI (bubbletea + lipgloss) | Done |
| HAR file replay & convert | Planned |
| Reporting — JSON, HTML, CSV | Planned |
| Variable & template system | Planned |
| Assertions & validation rules | Planned |

---

## Project Structure

```
loadforge/
├── cmd/loadforge/main.go          # Binary entrypoint
├── internal/
│   ├── cli/                       # CLI commands (cobra)
│   │   ├── root.go
│   │   ├── run.go                 # `run` command — launches test + UI
│   │   ├── validate.go
│   │   ├── har.go
│   │   └── version.go
│   ├── config/
│   │   └── config.go              # YAML config structs, Load(), Validate()
│   ├── engine/
│   │   ├── engine.go              # ExecuteStep() — runs a single HTTP step
│   │   ├── client.go              # HTTP client builder (HTTP/2, TLS, redirects)
│   │   └── builder.go             # Request builder (URL, body, auth)
│   ├── loader/
│   │   ├── loader.go              # Run() — entry point, broadcasts metrics to UI
│   │   ├── worker.go              # Worker goroutine — scenario + step loop
│   │   ├── collector.go           # Atomic stats collector, percentile calculation
│   │   ├── matrics.go             # Metrics struct, RPS, snapshots
│   │   ├── profile_constant.go    # Fixed N workers for the full duration
│   │   ├── profile_ramp.go        # Gradually increase workers over time
│   │   ├── profile_step.go        # Add workers in defined steps at intervals
│   │   └── profile_spike.go       # Base workers + periodic spike bursts
│   └── ui/
│       ├── model.go               # bubbletea Model — state, channels, RPS history
│       ├── ui.go                  # Init/Update/View, Run(), message types
│       ├── header.go              # Header bar — name, profile badge, progress bar
│       ├── stats.go               # Statistics panel — requests, RPS, error rate
│       ├── latency.go             # Latency panel — p50/p90/p95/p99 bar chart
│       └── sparkline.go           # RPS sparkline — live timeline chart
├── testdata/
│   └── sample.yaml                # Example test config
├── feature.md                     # Full feature roadmap
├── Makefile
└── go.mod
```

---

## Prerequisites

- Go 1.21 or later

---

## Build & Run

**Build the binary:**

```bash
make build
# Output: ./bin/loadforge
```

**Run without building:**

```bash
make run ARGS="validate testdata/sample.yaml"
```

**Or use `go run` directly:**

```bash
go run ./cmd/loadforge validate testdata/sample.yaml
go run ./cmd/loadforge version
```

---

## CLI Commands

### `run`

Runs a load test from a scenario YAML file. By default it launches the real-time terminal UI.

```bash
./bin/loadforge run testdata/sample.yaml
```

**Override config from the command line:**

```bash
# Override workers and duration
./bin/loadforge run scenario.yaml --workers 50 --duration 2m

# Disable the terminal UI and print plain text output instead
./bin/loadforge run scenario.yaml --no-ui

# Specify an output report file (for future reporting support)
./bin/loadforge run scenario.yaml --output report.html

# Pass variables
./bin/loadforge run scenario.yaml --var base_url=https://staging.example.com

# Load variables from a .env file
./bin/loadforge run scenario.yaml --env-file staging.env
```

**Available flags:**

| Flag | Short | Description |
|---|---|---|
| `--workers` | `-w` | Override number of concurrent workers |
| `--duration` | `-d` | Override test duration (e.g. `30s`, `2m`, `1h`) |
| `--output` | `-o` | Output report file path |
| `--var` | | Set a variable (`--var key=value`, repeatable) |
| `--env-file` | | Path to a `.env` file to load variables from |
| `--no-ui` | | Disable terminal UI, print plain text summary instead |

---

### `validate`

Parses and validates a scenario YAML file without running any requests. Useful for catching config errors early.

```bash
./bin/loadforge validate testdata/sample.yaml
```

Example output:

```
  Name     : Sample API Test
  Base URL : https://httpbin.org
  Scenarios: 1

  Scenario[1]: basic requests
    Weight : 0
    Steps  : 6
    Step[1]: GET request                    GET /get
    Step[2]: POST with JSON body            POST /post
    Step[3]: POST with form body            POST /post
    Step[4]: Bearer token auth              GET /bearer
    Step[5]: Basic auth                     GET /basic-auth/user/pass
    Step[6]: Custom auth header             GET /get

Config is valid.
```

### `version`

```bash
./bin/loadforge version
# loadforge v0.1.0
```

### `har`

HAR file replay and conversion commands are registered but not yet implemented.

```bash
./bin/loadforge har reply capture.har     # planned
./bin/loadforge har conver capture.har    # planned
```

---

## Terminal UI

When you run `loadforge run`, a full-screen terminal dashboard launches automatically (requires a terminal that supports ANSI colors).

The dashboard is built with [bubbletea](https://github.com/charmbracelet/bubbletea) and [lipgloss](https://github.com/charmbracelet/lipgloss) and refreshes every 250ms.

### What you see

```
loadforge  CONSTANT  my-scenario -> https://api.example.com
  [████████████████████░░░░░░░░░░░░░░░░░░░░]  24s / 60s  40%
──────────────────────────────────────────────────────────────
╭─────────────────────╮  ╭──────────────────────────────────╮
│ Statistics          │  │ Latency Percentiles               │
│                     │  │                                   │
│ Requests:      4821 │  │ p50  ████████████░░░░░░░  112ms  │
│ Successful:    4810 │  │ p90  ████████████████░░░░  198ms  │
│ Failed:           11│  │ p95  ██████████████████░░  231ms  │
│ Error Rate:   0.23% │  │ p99  ████████████████████  289ms  │
│ Avg RPS:     200.87 │  │                                   │
│ Data Received: 4.2MB│  ╰──────────────────────────────────╯
│ Elapsed:         24s│
│ Workers:          20│
╰─────────────────────╯
╭──────────────────────────────────────────────────────────────╮
│ RPS Over Time                                                 │
│ ▁▁▂▃▄▄▅▆▆▇▇▇████████████████████████████████████████████    │
│   current: 203.4 req/s   max: 218.1 req/s                    │
╰──────────────────────────────────────────────────────────────╯

  q quit    e toggle errors
```

### Panels

**Header** — scenario name, base URL, load profile badge, and a live progress bar showing elapsed vs total duration.

**Statistics** — total requests, successes, failures, error rate (color-coded green/yellow/red), avg RPS, total data received, elapsed time, and active worker count.

**Latency Percentiles** — horizontal bar chart for p50, p90, p95, p99. Bars are green under 200ms, yellow under 500ms, red above 500ms.

**RPS Sparkline** — rolling timeline of requests-per-second over the last 60 ticks, showing current and peak RPS.

**Errors panel** — appears automatically when failures occur, showing the top 3 error messages with counts.

### Keyboard controls

| Key | Action |
|---|---|
| `q` | Gracefully stop the test and exit |
| `Ctrl+C` | Same as `q` |
| `e` | Toggle the errors panel |

### Plain text mode

Use `--no-ui` to skip the dashboard and print a summary table after the test completes:

```
--------- RESULTS ---------
Total Requests    5000
Successful        4987
Failed            13
Error Rate        0.26%
Total Data        12.3 MB
Duration          30s
Avg RPS           166.67

--- Latency Percentiles ---
p50   112ms
p90   198ms
p95   231ms
p99   289ms

--- Status Codes ---
HTTP 200   4987
HTTP 500   13
```

---

## Writing a Scenario YAML

### Minimal example

```yaml
name: My API Test
base_url: https://api.example.com

scenarios:
  - name: health check
    steps:
      - name: ping
        method: GET
        url: /health

load:
  profile: constant
  workers: 10
  duration: 30s
```

### Full example with all step options

```yaml
name: Full API Test
base_url: https://httpbin.org

scenarios:
  - name: authenticated flow
    weight: 70
    steps:
      - name: login
        method: POST
        url: /post
        body:
          json:
            username: testuser
            password: secret
        options:
          timeout: 5s
          follow_redirects: true
          tls_skip_verify: false
          http2: true
        think: 500ms

      - name: get profile
        method: GET
        url: /get
        auth:
          bearer: my-token
        headers:
          X-Request-ID: loadforge-1

  - name: anonymous browse
    weight: 30
    steps:
      - name: browse
        method: GET
        url: /get

load:
  profile: ramp
  duration: 2m
  ramp_up:
    start_workers: 5
    end_workers: 50
    duration: 30s
```

---

## Load Profiles

### `constant`

Launches a fixed number of workers immediately and runs them for the full duration.

```yaml
load:
  profile: constant
  workers: 20
  duration: 1m
```

### `ramp`

Gradually increases from `start_workers` to `end_workers` over the `ramp_up.duration`, then holds until the total `duration` is reached.

```yaml
load:
  profile: ramp
  duration: 2m
  ramp_up:
    start_workers: 0
    end_workers: 50
    duration: 30s
```

### `step`

Starts at `start_workers` and adds `step_size` workers every `step_duration` until `max_workers` is reached.

```yaml
load:
  profile: step
  duration: 3m
  step:
    start_workers: 10
    step_size: 10
    step_duration: 30s
    max_workers: 100
```

### `spike`

Runs `base_workers` continuously. Every `spike_every`, it adds `spike_workers` for `spike_duration` then drops back to baseline. Tests recovery behavior.

```yaml
load:
  profile: spike
  duration: 2m
  spike:
    base_workers: 5
    spike_workers: 50
    spike_duration: 10s
    spike_every: 30s
```

---

## Auth Options

```yaml
# Bearer token
auth:
  bearer: your-token-here

# Basic auth
auth:
  basic:
    username: user
    password: pass

# Custom header
auth:
  header:
    key: X-API-Key
    value: your-api-key
```

---

## Request Body Options

```yaml
# JSON body
body:
  json:
    key: value

# Form encoded
body:
  form:
    field1: value1
    field2: value2

# Raw string
body:
  raw: "plain text body"
```

---

## Makefile Targets

| Target | Description |
|---|---|
| `make build` | Build binary to `./bin/loadforge` |
| `make run ARGS="..."` | Run via `go run` with arguments |
| `make tidy` | Run `go mod tidy` |
| `make clean` | Remove `./bin` directory |

---

## Planned Features

See [feature.md](feature.md) for the full roadmap, including:

- Per-request assertions (status code, body content, response time thresholds)
- Variable & template system (`{{.WorkerID}}`, `{{.UUID}}`, `--var key=value` CLI flags)
- HAR file replay from Chrome DevTools or Burp Suite exports
- Reports exported to JSON, HTML (with charts), and CSV
- Prometheus metrics endpoint during test runs
- Webhook notifications on test completion
