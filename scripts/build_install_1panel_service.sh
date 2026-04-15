#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

INSTALL_DIR="/opt"
CONFIG_DIR="/opt/1panel/conf"
BIN_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"

PANEL_PORT="9999"
AGENT_PORT="9998"
BIND_ADDRESS="0.0.0.0"
PANEL_MODE="dev"
PANEL_VERSION="v2.0.0"
PANEL_USERNAME="admin"
PANEL_PASSWORD="admin123"
PANEL_LANGUAGE="zh"
PANEL_EDITION="cn"
PANEL_ENTRANCE=""
SSL_MODE="disable"
IPV6_MODE="Disable"

SKIP_FRONTEND="false"
SKIP_NPM_INSTALL="false"
SKIP_ENABLE="false"
SKIP_START="false"

usage() {
    cat <<'EOF'
Usage:
  sudo ./scripts/build_install_1panel_service.sh [options]

Options:
  --install-dir DIR         Runtime install dir recorded into config and 1pctl. Default: /opt
  --panel-port PORT         1Panel core HTTP port. Default: 9999
  --agent-port PORT         1Panel agent port for non-master nodes. Default: 9998
  --bind-address ADDR       Core bind address. Default: 0.0.0.0
  --username NAME           Initial username in config. Default: admin
  --password PASS           Initial password in config. Default: admin123
  --language LANG           Language written into 1pctl. Default: zh
  --edition cn|intl         Panel edition. Default: cn
  --entrance PATH           Optional security entrance path
  --ssl disable|Enable|Mux  SSL mode written into config. Default: disable
  --skip-frontend           Skip frontend build
  --skip-npm-install        Skip npm install, use existing frontend/node_modules
  --skip-enable             Do not enable systemd services
  --skip-start              Do not restart/start services after install
  -h, --help                Show help

Examples:
  sudo ./scripts/build_install_1panel_service.sh

  sudo ./scripts/build_install_1panel_service.sh \
    --panel-port 8080 \
    --username admin \
    --password 'StrongPass123'
EOF
}

die() {
    printf '%s\n' "$*" >&2
    exit 1
}

log() {
    printf '[1Panel BuildInstall] %s\n' "$*"
}

require_root() {
    if [[ "${EUID}" -ne 0 ]]; then
        die "please run as root"
    fi
}

run_cmd() {
    log "$*"
    "$@"
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --install-dir) INSTALL_DIR="$2"; shift 2 ;;
        --panel-port) PANEL_PORT="$2"; shift 2 ;;
        --agent-port) AGENT_PORT="$2"; shift 2 ;;
        --bind-address) BIND_ADDRESS="$2"; shift 2 ;;
        --username) PANEL_USERNAME="$2"; shift 2 ;;
        --password) PANEL_PASSWORD="$2"; shift 2 ;;
        --language) PANEL_LANGUAGE="$2"; shift 2 ;;
        --edition) PANEL_EDITION="$2"; shift 2 ;;
        --entrance) PANEL_ENTRANCE="$2"; shift 2 ;;
        --ssl) SSL_MODE="$2"; shift 2 ;;
        --skip-frontend) SKIP_FRONTEND="true"; shift ;;
        --skip-npm-install) SKIP_NPM_INSTALL="true"; shift ;;
        --skip-enable) SKIP_ENABLE="true"; shift ;;
        --skip-start) SKIP_START="true"; shift ;;
        -h|--help) usage; exit 0 ;;
        *) die "unknown argument: $1" ;;
    esac
done

require_root

write_config() {
    mkdir -p "${CONFIG_DIR}"
    cat > "${CONFIG_DIR}/app.yaml" <<EOF
base:
  install_dir: ${INSTALL_DIR}
  mode: ${PANEL_MODE}
  is_demo: false
  is_offline: false
  is_fxplay: false
  port: ${AGENT_PORT}
  username: ${PANEL_USERNAME}
  password: ${PANEL_PASSWORD}
  version: ${PANEL_VERSION}
  edition: ${PANEL_EDITION}

conn:
  port: ${PANEL_PORT}
  bindAddress: ${BIND_ADDRESS}
  ipv6: ${IPV6_MODE}
  ssl: ${SSL_MODE}
  entrance: ${PANEL_ENTRANCE}

log:
  level: debug
  time_zone: Asia/Shanghai
  log_name: 1Panel
  log_suffix: .log
  max_backup: 10
EOF
}

write_1pctl() {
    cat > "${BIN_DIR}/1pctl" <<EOF
#!/usr/bin/env bash
BASE_DIR=${INSTALL_DIR}
ORIGINAL_PORT=${PANEL_PORT}
ORIGINAL_VERSION=${PANEL_VERSION}
ORIGINAL_USERNAME=${PANEL_USERNAME}
ORIGINAL_PASSWORD=${PANEL_PASSWORD}
ORIGINAL_ENTRANCE=${PANEL_ENTRANCE}
LANGUAGE=${PANEL_LANGUAGE}
PANEL_EDITION=${PANEL_EDITION}
CHANGE_USER_INFO=

if [[ \$# -eq 0 ]]; then
  exec "${BIN_DIR}/1panel-core"
fi

exec "${BIN_DIR}/1panel-core" "\$@"
EOF
    chmod 755 "${BIN_DIR}/1pctl"
}

write_service_unit() {
    local name="$1"
    local exec_path="$2"
    cat > "${SERVICE_DIR}/${name}.service" <<EOF
[Unit]
Description=${name}
After=network.target docker.service
Wants=network.target

[Service]
Type=simple
ExecStart=${exec_path}
Restart=always
RestartSec=3
WorkingDirectory=${INSTALL_DIR}
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF
}

build_frontend() {
    if [[ "${SKIP_FRONTEND}" == "true" ]]; then
        log "skip frontend build"
        return 0
    fi
    if [[ "${SKIP_NPM_INSTALL}" == "true" ]]; then
        run_cmd bash -lc "cd \"${REPO_DIR}/frontend\" && npm run build:pro"
    else
        run_cmd bash -lc "cd \"${REPO_DIR}/frontend\" && npm install && npm run build:pro"
    fi
}

build_binaries() {
    mkdir -p "${REPO_DIR}/build"
    run_cmd bash -lc "cd \"${REPO_DIR}/core\" && CGO_ENABLED=0 GOOS=linux GOARCH=\$(go env GOARCH) go build -trimpath -ldflags '-s -w' -o \"${REPO_DIR}/build/1panel-core\" ./cmd/server/main.go"
    run_cmd bash -lc "cd \"${REPO_DIR}/agent\" && CGO_ENABLED=0 GOOS=linux GOARCH=\$(go env GOARCH) go build -trimpath -ldflags '-s -w' -o \"${REPO_DIR}/build/1panel-agent\" ./cmd/server/main.go"
}

install_binaries() {
    install -m 755 "${REPO_DIR}/build/1panel-core" "${BIN_DIR}/1panel-core"
    install -m 755 "${REPO_DIR}/build/1panel-agent" "${BIN_DIR}/1panel-agent"
}

ensure_runtime_dirs() {
    mkdir -p "${INSTALL_DIR}/1panel"
    mkdir -p "${INSTALL_DIR}/1panel/db"
    mkdir -p "${INSTALL_DIR}/1panel/log"
    mkdir -p "${INSTALL_DIR}/1panel/tmp"
}

reload_systemd() {
    run_cmd systemctl daemon-reload
    if [[ "${SKIP_ENABLE}" != "true" ]]; then
        run_cmd systemctl enable 1panel-core.service
        run_cmd systemctl enable 1panel-agent.service
    fi
    if [[ "${SKIP_START}" != "true" ]]; then
        run_cmd systemctl restart 1panel-core.service
        run_cmd systemctl restart 1panel-agent.service
    fi
}

main() {
    log "repo dir: ${REPO_DIR}"
    build_frontend
    build_binaries
    ensure_runtime_dirs
    install_binaries
    write_config
    write_1pctl
    write_service_unit "1panel-core" "${BIN_DIR}/1panel-core"
    write_service_unit "1panel-agent" "${BIN_DIR}/1panel-agent"
    reload_systemd
    log "done"
    log "config: ${CONFIG_DIR}/app.yaml"
    log "core binary: ${BIN_DIR}/1panel-core"
    log "agent binary: ${BIN_DIR}/1panel-agent"
}

main "$@"
