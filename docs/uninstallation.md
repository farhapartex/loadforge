# Uninstallation

```bash
sudo loadforge --uninstall
```

This single command:

1. Stops and removes the background service (systemd on Linux, launchd on macOS)
2. Removes `/usr/local/bin/loadforge` and `/usr/local/bin/loadforge-web`
3. Deletes `~/.loadforge/` including config, run history, and logs
