# Scenario File

LoadForge scenarios are defined in YAML. A scenario file describes the requests to send, how to authenticate, and how much load to apply.

## Full reference

```yaml
name: Test Name
base_url: https://api.example.com

scenarios:
  - name: Scenario Name
    weight: 100              # Relative probability when multiple scenarios are defined
    steps:
      - name: Step Name
        method: GET           # GET, POST, PUT, PATCH, DELETE, HEAD
        url: /path            # Relative (uses base_url) or absolute URL

        headers:
          X-Custom-Header: value

        body:
          raw: "raw string body"
          json:
            key: value
          form:
            field: value

        auth:
          basic:
            username: user
            password: pass
          bearer: your-token
          header:
            key: Authorization
            value: Bearer your-token

        options:
          timeout: 10s
          follow_redirects: true
          tls_skip_verify: false
          http2: true

        think: 100ms           # Pause after this step

load:
  profile: constant            # constant | ramp | step | spike
  workers: 10
  duration: 30s

assertions:
  - metric: p95_latency
    operator: less_than
    value: 500
    enabled: true
```

## Multiple scenarios

When multiple scenarios are defined, workers pick one at random weighted by the `weight` field.

```yaml
scenarios:
  - name: Read path
    weight: 80
    steps:
      - method: GET
        url: /items

  - name: Write path
    weight: 20
    steps:
      - method: POST
        url: /items
        body:
          json:
            name: test
```

## Body types

Only one body type is used per step. Priority: `json` → `form` → `raw`.

| Type | Content-Type set automatically |
|------|-------------------------------|
| `json` | `application/json` |
| `form` | `application/x-www-form-urlencoded` |
| `raw` | none (set manually via `headers`) |
