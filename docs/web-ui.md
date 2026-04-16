# Web UI

The web UI is a browser-based dashboard for running and reviewing load tests without using the CLI.

## Accessing the UI

After installation the web UI starts automatically as a background service. Open [http://localhost:8090](http://localhost:8090).

Default credentials: **admin / admin**

Change the credentials in `http://localhost:8090/settings`

---

## Running a test

1. Paste an OpenAPI or Swagger spec URL into the input field, or select a saved scenario
2. Set load parameters: profile, workers, duration
3. Click **Start Test**

Live logs stream to the page via Server-Sent Events as the test runs.

---

## History

Every completed run is saved and can be found in this page: to `http://localhost:8090/history`. The **History** page lists all past runs with:

- Latency percentiles (P50, P90, P95, P99)
- RPS, error rate, success rate
- Status code distribution
- Assertion pass/fail results

