# CLI Reference

## loadforge run

Run a load test from a scenario file.

```bash
loadforge run [scenario.yaml] [flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--workers` | `-w` | Override worker count from config |
| `--duration` | `-d` | Override duration (e.g. `30s`, `2m`) |
| `--output` | `-o` | Save report to a JSON file |
| `--var` | | Set a variable (`key=value`) |
| `--env-file` | | Load variables from a `.env` file |
| `--no-ui` | | Plain text output instead of the terminal UI |
| `--verbose` | `-v` | Enable verbose logging |

### Examples

```bash
# Run with defaults from the file
loadforge run scenario.yaml

# Override workers and duration
loadforge run scenario.yaml --workers 50 --duration 2m

# Inject a variable
loadforge run scenario.yaml --var BASE_URL=https://staging.example.com

# Load variables from .env
loadforge run scenario.yaml --env-file .env
```

---

## loadforge validate

Validate a scenario file without running it.

```bash
loadforge validate [scenario.yaml]
```

---

## loadforge version

Print the installed version.

```bash
loadforge version
```

---

## loadforge --uninstall

Stop the web service and remove all LoadForge files from the system.

```bash
sudo loadforge --uninstall
```

This removes:

- `/usr/local/bin/loadforge`
- `/usr/local/bin/loadforge-web`
- `~/.loadforge/` (config, history, logs)
- The background service (systemd or launchd)
