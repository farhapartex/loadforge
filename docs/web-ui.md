# Web UI

The web UI is a browser-based dashboard for running and reviewing load tests without using the CLI.

## Accessing the UI

After installation the web UI starts automatically as a background service. Open [http://localhost:8090](http://localhost:8090).

Default credentials: **admin / admin**

Change the credentials in `~/.loadforge/web.yml`:

```yaml
username: yourname
password: yourpassword
```

Then restart the service:

=== "macOS"
    ```bash
    sudo launchctl unload /Library/LaunchDaemons/com.loadforge.web.plist
    sudo launchctl load /Library/LaunchDaemons/com.loadforge.web.plist
    ```

=== "Linux"
    ```bash
    sudo systemctl restart loadforge-web
    ```

---

## Running a test

1. Paste an OpenAPI or Swagger spec URL into the input field, or select a saved scenario
2. Set load parameters: profile, workers, duration
3. Click **Start Test**

Live logs stream to the page via Server-Sent Events as the test runs.

---

## History

Every completed run is saved to `~/.loadforge/load_forge_history.json` (up to 100 entries). The **History** page lists all past runs with:

- Latency percentiles (P50, P90, P95, P99)
- RPS, error rate, success rate
- Status code distribution
- Assertion pass/fail results

---

## Configuration file

The web server reads `~/.loadforge/web.yml`:

```yaml
addr: :8090
username: admin
password: admin
session_ttl: 24h
log_file: /home/user/.loadforge/loadforge.logs
history_file: /home/user/.loadforge/load_forge_history.json

assertions:
  - metric: p95_latency
    operator: less_than
    value: 500
    enabled: true
```
