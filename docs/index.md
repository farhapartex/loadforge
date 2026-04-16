# LoadForge

> Developer-first HTTP load testing — run from the terminal or a browser-based UI.

LoadForge is an open-source HTTP load testing tool built for developers and QA engineers who want to stress-test their APIs without fighting complex configuration or expensive SaaS platforms.

Write a simple YAML scenario, point it at your API, and LoadForge hammers it with concurrent workers while giving you live metrics in a clean terminal UI or a browser dashboard.

---

## Why LoadForge?

Most load testing tools are either too heavy (JMeter, Gatling) or too limited (ab, wrk). LoadForge sits in the middle:

- **Zero dependencies** — a single binary, no JVM, no Node, no Docker required
- **Two interfaces** — use the CLI for scripts and CI pipelines, or the web UI for interactive testing
- **OpenAPI / Swagger support** — point at an API spec URL and LoadForge generates and runs the scenario for you
- **Multiple load profiles** — constant, ramp, step, and spike traffic patterns out of the box
- **Assertions & SLA thresholds** — define pass/fail criteria and catch regressions automatically
- **History tracking** — every run is persisted so you can compare results over time

---

## Features

| Feature | Description |
|---|---|
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

## Get started

```bash
curl -fsSL https://github.com/farhapartex/loadforge/releases/latest/download/install.sh | sudo bash
```

Then open [http://localhost:8090](http://localhost:8090) — the web UI starts automatically.

See the [Installation](installation.md) guide for full details.

---

## Video Walkthrough

New to LoadForge? Watch the full demo to see installation, running your first test, and reviewing results — all in under 4 minutes.

[![LoadForge Demo](https://img.youtube.com/vi/Z8Q48Vo-BEk/0.jpg)](https://www.youtube.com/watch?v=Z8Q48Vo-BEk)
