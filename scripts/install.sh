#!/usr/bin/env bash
set -euo pipefail

REPO="farhapartex/loadforge"
BINARY="loadforge"
INSTALL_DIR="/usr/local/bin"
APP_DIR_NAME=".loadforge"

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
    # macOS fallback
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
  local version
  version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  [ -n "$version" ] || error "Could not determine the latest release version."
  echo "$version"
}

write_default_config() {
  local config_file="$1"
  local history_file="$2"
  local log_file="$3"

  cat > "$config_file" <<EOF
addr: :8080
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
  local user
  local home
  local app_dir
  local config_file
  local history_file
  local log_file

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

  # Ensure the real user owns the directory (not root)
  if [ -n "${SUDO_USER:-}" ]; then
    chown -R "${user}:$(id -gn "$user")" "$app_dir"
  fi
}

install() {
  require curl
  require tar

  local os arch version ver_num archive_name download_url tmpdir

  os=$(detect_os)
  arch=$(detect_arch)
  version=$(latest_version)
  ver_num="${version#v}"

  archive_name="${BINARY}_${ver_num}_${os}_${arch}.tar.gz"
  download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"

  info "Installing ${BINARY} ${version} (${os}/${arch})"
  info "Downloading ${download_url}"

  tmpdir=$(mktemp -d)
  trap 'rm -rf "$tmpdir"' EXIT

  curl -fsSL "$download_url" -o "${tmpdir}/${archive_name}" \
    || error "Download failed. Check that ${version} exists at https://github.com/${REPO}/releases"

  tar -xzf "${tmpdir}/${archive_name}" -C "$tmpdir"

  install -m 755 "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}" \
    || error "Failed to install binary to ${INSTALL_DIR}. Try running with sudo."

  install -m 755 "${tmpdir}/${BINARY}-web" "${INSTALL_DIR}/${BINARY}-web" \
    || error "Failed to install loadforge-web to ${INSTALL_DIR}. Try running with sudo."

  info "Installed binary to ${INSTALL_DIR}/${BINARY}"
  info "Installed binary to ${INSTALL_DIR}/${BINARY}-web"

  setup_app_dir

  echo ""
  info "Installation complete!"
  info "  Binary  : ${INSTALL_DIR}/${BINARY}"
  info "  Config  : $(real_home)/${APP_DIR_NAME}/web.yml"
  info "  History : $(real_home)/${APP_DIR_NAME}/load_forge_history.json"
  echo ""
  info "Run '${BINARY} version' to verify."
  info "Run '${BINARY} --help' to get started."
}

uninstall() {
  local target="${INSTALL_DIR}/${BINARY}"
  local app_dir
  app_dir="$(real_home)/${APP_DIR_NAME}"

  if [ ! -f "$target" ]; then
    error "${BINARY} is not installed at ${target}."
  fi

  rm -f "$target"
  rm -f "${INSTALL_DIR}/${BINARY}-web"
  info "Removed ${target}"
  info "Removed ${INSTALL_DIR}/${BINARY}-web"

  if [ -d "$app_dir" ]; then
    rm -rf "$app_dir"
    info "Removed ${app_dir}"
  fi

  info "${BINARY} has been fully uninstalled."
}

case "${1:-install}" in
  install)   install ;;
  uninstall) uninstall ;;
  *) error "Unknown command '${1}'. Use: install | uninstall" ;;
esac
