loadforge — Feature Design
Core Philosophy

Single binary, zero dependencies to install
Config-driven (YAML) so tests are repeatable and version-controllable
Real-time terminal UI, not just a final report
Built for developers and platform engineers, not just QA


Feature 1 — Request Engine

Support HTTP/1.1 and HTTP/2
Methods: GET, POST, PUT, PATCH, DELETE, HEAD
Custom headers per request
Request body — raw, JSON, form-encoded
Basic Auth, Bearer Token, and custom auth headers
Per-request timeout configuration
Follow or block redirects (configurable)
TLS skip verify option (for internal/staging environments)

Feature 2 — Load Profiles

Constant — fixed N workers for the whole duration
Ramp-up — gradually increase workers from 0 to N over a defined period
Step — increase load in steps (e.g. 10 → 20 → 50 → 100 users, every 30s)
Spike — sudden burst then back to baseline, to test recovery
Duration-based runs (e.g. run for 2 minutes) or request-count-based (e.g. send 10,000 requests total)

Feature 3 — Scripted Scenarios
This is what separates loadforge from simple tools like hey or wrk.

A scenario is a sequence of HTTP steps executed as one virtual user flow
Each step can use the response of a previous step (e.g. login → extract token → use token in next request)
Variable extraction from responses — JSON path, regex, header value
Conditional steps — skip a step based on previous response status
Per-step think time / delay (to simulate real user behavior)
Multiple scenarios in one config with weighted distribution (e.g. 70% browse, 30% checkout)

Feature 4 — Assertions & Validation

Assert response status code (exact or range like 2xx)
Assert response body contains a string or matches a regex
Assert response JSON field equals an expected value
Assert response time is below a threshold
Failed assertions are tracked separately from network errors
Option to stop the test if assertion failure rate exceeds a threshold


Feature 5 — Real-time Terminal UI
Built with bubbletea + lipgloss:

Live updating dashboard showing:

Requests/sec (current and average)
Active workers / goroutines
Total requests sent, succeeded, failed
Latency percentiles — p50, p90, p95, p99
Error rate %
A live latency histogram (bar chart in terminal)
Timeline sparkline of RPS over time


Color-coded status (green / yellow / red thresholds)
Press q to gracefully stop the test early

Feature 6 — HAR File Replay

Import a .har file (exported from Chrome DevTools / Burp Suite)
Parse all requests from the HAR
Replay them as a scenario in order with original headers/bodies
Option to parameterize replayed requests (replace hardcoded values with variables)


Feature 7 — Reporting

Final summary printed to terminal after test completes
Export results to:

JSON — full raw results, machine readable
HTML — self-contained report with charts (using embedded JS charting)
CSV — per-request log for external analysis


Report includes: latency distribution, error breakdown, throughput over time, assertion failures


Feature 8 — Variable & Environment System

Global variables defined in config or passed via CLI flags (--var key=value)
Environment files (.env style) loadable via --env-file
Built-in variables: {{.Timestamp}}, {{.WorkerID}}, {{.Iteration}}, {{.RandomInt}}, {{.UUID}}
Template syntax using Go's text/template for all string fields in config


Feature 9 — CLI Interface
loadforge run scenario.yaml
loadforge run scenario.yaml --workers 50 --duration 2m
loadforge run scenario.yaml --output report.html
loadforge run scenario.yaml --var base_url=https://staging.example.com
loadforge har replay capture.har
loadforge har convert capture.har --output scenario.yaml
loadforge validate scenario.yaml      # dry-run, checks config is valid
loadforge version


Feature 10 — Observability Hooks

Push live metrics to Prometheus (expose /metrics endpoint during the run)
Send results to InfluxDB for Grafana dashboards
Webhook notification on test complete (Slack, Teams, custom)
Optional output of per-request trace log to a file