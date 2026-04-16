#!/usr/bin/env bash
set -euo pipefail

REPO="farhapartex/loadforge"
BINARY="loadforge"
INSTALL_DIR="/usr/local/bin"
APP_DIR_NAME=".loadforge"
LAUNCHD_LABEL="com.loadforge.web"
LAUNCHD_PLIST="/Library/LaunchDaemons/${LAUNCHD_LABEL}.plist"
SYSTEMD_SERVICE="/etc/systemd/system/loadforge-web.service"

info()  { echo "[loadforge] $*"; }
error() { echo "[loadforge] ERROR: $*" >&2; exit 1; }

require() {
  command -v "$1" >/dev/null 2>&1 || error "'$1' is required but not installed."
}

real_user() {
  if [ -n "${SUDO_USER:-}" ]; then
    echo "$SUDO_USER"
  else
    echo "$(whoami)"
  fi
}

real_home() {
  local user
  user=$(real_user)
  if command -v getent >/dev/null 2>&1; then
    getent passwd "$user" | cut -d: -f6
  else
    eval echo "~$user"
  fi
}

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux"  ;;
    Darwin*) echo "darwin" ;;
    *) error "Unsupported OS: $(uname -s). Only Linux and macOS are supported." ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) error "Unsupported architecture: $(uname -m)." ;;
  esac
}

latest_version() {
  local url version
  url=$(curl -fsSL -o /dev/null -w "%{url_effective}" "https://github.com/${REPO}/releases/latest")
  version="${url##*/}"
  [ -n "$version" ] || error "Could not determine the latest release version."
  echo "$version"
}

write_default_config() {
  local config_file="$1"
  local history_file="$2"
  local log_file="$3"

  cat > "$config_file" <<EOF
addr: :8090
username: admin
password: admin
session_ttl: 24h
log_file: ${log_file}
history_file: ${history_file}
assertions:
    - metric: p95_latency
      operator: less_than
      value: 500
      enabled: true
    - metric: p99_latency
      operator: less_than
      value: 1000
      enabled: true
    - metric: error_rate
      operator: less_than
      value: 1
      enabled: true
    - metric: success_rate
      operator: greater_than_or_equal
      value: 99
      enabled: false
    - metric: rps
      operator: greater_than
      value: 10
      enabled: false
EOF
}

setup_app_dir() {
  local user home app_dir config_file history_file log_file

  user=$(real_user)
  home=$(real_home)
  app_dir="${home}/${APP_DIR_NAME}"
  config_file="${app_dir}/web.yml"
  history_file="${app_dir}/load_forge_history.json"
  log_file="${app_dir}/load_forge.logs"

  info "Setting up app directory at ${app_dir}"
  mkdir -p "$app_dir"

  if [ ! -f "$config_file" ]; then
    write_default_config "$config_file" "$history_file" "$log_file"
    info "Created default config at ${config_file}"
  else
    info "Config already exists at ${config_file} — skipping"
  fi

  if [ ! -f "$history_file" ]; then
    echo "[]" > "$history_file"
    info "Created history file at ${history_file}"
  fi

  if [ -n "${SUDO_USER:-}" ]; then
    chown -R "${user}:$(id -gn "$user")" "$app_dir"
  fi
}

setup_service() {
  local os user home app_dir log_file config_file

  os=$(detect_os)
  user=$(real_user)
  home=$(real_home)
  app_dir="${home}/${APP_DIR_NAME}"
  config_file="${app_dir}/web.yml"
  log_file="${app_dir}/load_forge.logs"

  if [ "$os" = "linux" ]; then
    info "Registering systemd service..."
    cat > "$SYSTEMD_SERVICE" <<EOF
[Unit]
Description=LoadForge Web UI
After=network.target

[Service]
ExecStart=${INSTALL_DIR}/${BINARY}-web --config ${config_file}
Restart=always
User=${user}
StandardOutput=append:${log_file}
StandardError=append:${log_file}

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    systemctl enable loadforge-web
    systemctl start loadforge-web
    info "Service started. Check status: systemctl status loadforge-web"

  elif [ "$os" = "darwin" ]; then
    info "Registering launchd service..."
    cat > "$LAUNCHD_PLIST" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>${LAUNCHD_LABEL}</string>
    <key>ProgramArguments</key>
    <array>
        <string>${INSTALL_DIR}/${BINARY}-web</string>
        <string>--config</string>
        <string>${config_file}</string>
    </array>
    <key>UserName</key>
    <string>${user}</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>${log_file}</string>
    <key>StandardErrorPath</key>
    <string>${log_file}</string>
</dict>
</plist>
EOF
    launchctl load -w "$LAUNCHD_PLIST"
    info "Service started. Open http://localhost:8090 in your browser."
  fi
}

remove_service() {
  local os
  os=$(detect_os)

  if [ "$os" = "linux" ]; then
    if systemctl is-active --quiet loadforge-web 2>/dev/null; then
      systemctl stop loadforge-web
      info "Stopped loadforge-web service"
    fi
    if systemctl is-enabled --quiet loadforge-web 2>/dev/null; then
      systemctl disable loadforge-web
    fi
    if [ -f "$SYSTEMD_SERVICE" ]; then
      rm -f "$SYSTEMD_SERVICE"
      systemctl daemon-reload
      info "Removed systemd service"
    fi

  elif [ "$os" = "darwin" ]; then
    if [ -f "$LAUNCHD_PLIST" ]; then
      launchctl unload -w "$LAUNCHD_PLIST" 2>/dev/null || true
      rm -f "$LAUNCHD_PLIST"
      info "Removed launchd service"
    fi
  fi
}

do_install() {
  require curl

  local os arch version binary_name web_binary_name download_url web_download_url

  os=$(detect_os)
  arch=$(detect_arch)
  version=$(latest_version)

  binary_name="${BINARY}-${os}_${arch}"
  web_binary_name="${BINARY}-web-${os}_${arch}"
  download_url="https://github.com/${REPO}/releases/download/${version}/${binary_name}"
  web_download_url="https://github.com/${REPO}/releases/download/${version}/${web_binary_name}"

  info "Installing ${BINARY} ${version} (${os}/${arch})"
  info "Downloading ${download_url}"

  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT

  curl -fsSL "$download_url" -o "${tmpdir}/${BINARY}" \
    || error "Download failed. Check that ${version} exists at https://github.com/${REPO}/releases"

  curl -fsSL "$web_download_url" -o "${tmpdir}/${BINARY}-web" \
    || error "Download failed. Check that ${version} exists at https://github.com/${REPO}/releases"

  install -m 755 "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}" \
    || error "Failed to install binary to ${INSTALL_DIR}. Try running with sudo."

  install -m 755 "${tmpdir}/${BINARY}-web" "${INSTALL_DIR}/${BINARY}-web" \
    || error "Failed to install loadforge-web to ${INSTALL_DIR}. Try running with sudo."

  info "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
  info "Installed ${BINARY}-web to ${INSTALL_DIR}/${BINARY}-web"

  setup_app_dir
  setup_service

  echo ""
  info "Installation complete!"
  info "  Binary  : ${INSTALL_DIR}/${BINARY}"
  info "  Config  : $(real_home)/${APP_DIR_NAME}/web.yml"
  info "  History : $(real_home)/${APP_DIR_NAME}/load_forge_history.json"
  info "  Web UI  : http://localhost:8090  (admin / admin)"
  echo ""
  info "To uninstall: sudo loadforge --uninstall"
}

uninstall() {
  local app_dir
  app_dir="$(real_home)/${APP_DIR_NAME}"

  remove_service

  if [ -f "${INSTALL_DIR}/${BINARY}" ]; then
    rm -f "${INSTALL_DIR}/${BINARY}"
    info "Removed ${INSTALL_DIR}/${BINARY}"
  fi

  if [ -f "${INSTALL_DIR}/${BINARY}-web" ]; then
    rm -f "${INSTALL_DIR}/${BINARY}-web"
    info "Removed ${INSTALL_DIR}/${BINARY}-web"
  fi

  if [ -d "$app_dir" ]; then
    rm -rf "$app_dir"
    info "Removed ${app_dir}"
  fi

  info "LoadForge has been fully uninstalled."
}

case "${1:-install}" in
  install)   do_install ;;
  uninstall) uninstall ;;
  *) error "Unknown command '${1}'. Use: install | uninstall" ;;
esac
