# loadforge

HTTP load testing tool with a browser-based UI and CLI.

---

## Build

```bash
# CLI tool
make build
# → ./bin/loadforge

# Web UI server
make build-web
# → ./bin/loadforge-web
```

---

## Run

### Web UI

```bash
./bin/loadforge-web
```

Opens at `http://localhost:8080`. Default login: `admin` / `admin`.

To customise the address, credentials, or log file, create a `web.yml` alongside the binary:

```yaml
addr: ":8080"
username: "admin"
password: "admin"
session_ttl: "24h"
log_file: "load_forge.logs"
```

### CLI

```bash
./bin/loadforge run testdata/sample.yaml

# Override workers and duration
./bin/loadforge run testdata/sample.yaml --workers 20 --duration 1m

# Disable the terminal UI
./bin/loadforge run testdata/sample.yaml --no-ui

# Validate a scenario file without running it
./bin/loadforge validate testdata/sample.yaml
```

---

## Running a load test from the UI

1. Open `http://localhost:8080` and log in.
2. Click **Run New Test**.
3. Paste your **API Doc URL** — a publicly reachable OpenAPI 3.x or Swagger 2.0 spec (JSON or YAML).
4. Optionally enter a **JWT token** if the spec endpoint or the API under test requires one.
5. Set **workers**, **duration**, and **load profile**, then click **Start Test**.

The app fetches the spec, extracts all endpoints, generates a load test config, and starts running immediately. Live logs appear on the home page. Past runs are listed under **History**.

---

## Load profiles

| Profile | Description |
|---|---|
| `constant` | Fixed number of workers for the full duration |
| `ramp` | Gradually increase workers from start to end |
| `step` | Add workers in steps at set intervals |
| `spike` | Base workers with periodic traffic spikes |

---

## Makefile targets

| Target | Description |
|---|---|
| `make build` | Build CLI binary to `./bin/loadforge` |
| `make build-web` | Build web server to `./bin/loadforge-web` |
| `make run ARGS="..."` | Run CLI via `go run` |
| `make run-web` | Run web server via `go run` |
| `make tidy` | Run `go mod tidy` |
| `make clean` | Remove `./bin` |
