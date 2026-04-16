# Quick Start

## Web UI

After installation the web UI starts automatically. Open [http://localhost:8090](http://localhost:8090) in your browser.

Default credentials: **admin / admin**

1. Paste an OpenAPI or Swagger spec URL into the input field
2. Set your load parameters (workers, duration, profile)
3. Click **Start Test**

Live logs stream to the dashboard as the test runs. Past runs appear under **History**.

---

## CLI

### 1. Write a scenario file

```yaml title="scenario.yaml"
name: My API Test
base_url: https://api.example.com

scenarios:
  - name: Get users
    steps:
      - name: List users
        method: GET
        url: /users

load:
  profile: constant
  workers: 10
  duration: 30s
```

### 2. Run the test

```bash
loadforge run scenario.yaml
```

A real-time terminal UI shows live metrics. Press `q` to stop early.

### 3. Validate without running

```bash
loadforge validate scenario.yaml
```
