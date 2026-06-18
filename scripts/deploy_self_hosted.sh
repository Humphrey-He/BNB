#!/usr/bin/env bash

set -euo pipefail

require_env() {
    local key="$1"
    if [[ -z "${!key:-}" ]]; then
        echo "[FATAL] missing env: ${key}" >&2
        exit 1
    fi
}

log() {
    printf '[%s] %s\n' "$(date '+%F %T')" "$*"
}

run_sudo() {
    printf '%s\n' "${REMOTE_SUDO_PASSWORD}" | sudo -S -p '' "$@"
}

sync_dir() {
    local src="$1"
    local dest="$2"
    run_sudo mkdir -p "${dest}"
    if command -v rsync >/dev/null 2>&1; then
        run_sudo rsync -a --delete "${src}/" "${dest}/"
    else
        run_sudo rm -rf "${dest}"
        run_sudo mkdir -p "${dest}"
        run_sudo cp -a "${src}/." "${dest}/"
    fi
}

require_env GITHUB_WORKSPACE
require_env DEPLOY_ROOT
require_env FRONTEND_WEB_ROOT
require_env NGINX_CONF_PATH
require_env BNB_SERVER_NAME
require_env GO_API_PORT
require_env RUST_RPC_PORT
require_env NATS_URL
require_env APP_ENV
require_env LOG_LEVEL
require_env POSTGRES_HOST
require_env POSTGRES_PORT
require_env POSTGRES_DB
require_env POSTGRES_USER
require_env POSTGRES_PASSWORD
require_env REDIS_HOST
require_env REDIS_PORT
require_env REMOTE_SUDO_PASSWORD

WORKDIR="$(mktemp -d)"
trap 'rm -rf "${WORKDIR}"' EXIT

RELEASE_DIR="${GITHUB_WORKSPACE}/release"
API_SRC="${RELEASE_DIR}/api"
WORKERS_SRC="${RELEASE_DIR}/workers"
FRONTEND_SRC="${RELEASE_DIR}/frontend-dist"
NODE_SRC="${RELEASE_DIR}/node"

[[ -x "${API_SRC}/api-server-linux-amd64" ]] || { echo "[FATAL] missing Go binary" >&2; exit 1; }
[[ -x "${WORKERS_SRC}/scanner-linux-amd64" ]] || { echo "[FATAL] missing scanner binary" >&2; exit 1; }
[[ -x "${WORKERS_SRC}/parser-linux-amd64" ]] || { echo "[FATAL] missing parser binary" >&2; exit 1; }
[[ -x "${WORKERS_SRC}/confirm-worker-linux-amd64" ]] || { echo "[FATAL] missing confirm-worker binary" >&2; exit 1; }
[[ -x "${WORKERS_SRC}/ledger-service-linux-amd64" ]] || { echo "[FATAL] missing ledger-service binary" >&2; exit 1; }
[[ -x "${WORKERS_SRC}/rpc-health-linux-amd64" ]] || { echo "[FATAL] missing rpc-health binary" >&2; exit 1; }
[[ -f "${FRONTEND_SRC}/index.html" ]] || { echo "[FATAL] missing frontend dist" >&2; exit 1; }
[[ -x "${NODE_SRC}/verifiable-rust-chain-node" ]] || { echo "[FATAL] missing Rust binary" >&2; exit 1; }

API_DIR="${DEPLOY_ROOT}/api"
WORKERS_DIR="${DEPLOY_ROOT}/workers"
NODE_DIR="${DEPLOY_ROOT}/node"
RELEASES_DIR="${DEPLOY_ROOT}/releases/${GITHUB_SHA:-manual}"

log "Preparing target directories"
run_sudo mkdir -p "${API_DIR}" "${API_DIR}/scripts/migrations" "${WORKERS_DIR}" "${NODE_DIR}/target/release" "${RELEASES_DIR}"

log "Saving release metadata"
cat > "${WORKDIR}/release-meta.txt" <<EOF
commit=${GITHUB_SHA:-manual}
run_id=${GITHUB_RUN_ID:-manual}
run_attempt=${GITHUB_RUN_ATTEMPT:-manual}
deployed_at=$(date '+%F %T')
EOF
run_sudo mv "${WORKDIR}/release-meta.txt" "${RELEASES_DIR}/release-meta.txt"

log "Deploying frontend"
sync_dir "${FRONTEND_SRC}" "${FRONTEND_WEB_ROOT}"

log "Deploying Go API"
run_sudo install -m 755 "${API_SRC}/api-server-linux-amd64" "${API_DIR}/api-server-linux-amd64"
if [[ -d "${API_SRC}/migrations" ]]; then
    sync_dir "${API_SRC}/migrations" "${API_DIR}/scripts/migrations"
fi

log "Deploying worker binaries"
run_sudo install -m 755 "${WORKERS_SRC}/scanner-linux-amd64" "${WORKERS_DIR}/scanner-linux-amd64"
run_sudo install -m 755 "${WORKERS_SRC}/parser-linux-amd64" "${WORKERS_DIR}/parser-linux-amd64"
run_sudo install -m 755 "${WORKERS_SRC}/confirm-worker-linux-amd64" "${WORKERS_DIR}/confirm-worker-linux-amd64"
run_sudo install -m 755 "${WORKERS_SRC}/ledger-service-linux-amd64" "${WORKERS_DIR}/ledger-service-linux-amd64"
run_sudo install -m 755 "${WORKERS_SRC}/rpc-health-linux-amd64" "${WORKERS_DIR}/rpc-health-linux-amd64"

REDIS_URL="redis://${REDIS_HOST}:${REDIS_PORT}"
REDIS_PASSWORD_LINE="# REDIS_PASSWORD="
if [[ -n "${REDIS_PASSWORD:-}" ]]; then
    REDIS_URL="redis://:${REDIS_PASSWORD}@${REDIS_HOST}:${REDIS_PORT}"
    REDIS_PASSWORD_LINE="REDIS_PASSWORD=${REDIS_PASSWORD}"
fi

log "Writing Go API environment"
cat > "${WORKDIR}/api.env" <<EOF
POSTGRES_HOST=${POSTGRES_HOST}
POSTGRES_PORT=${POSTGRES_PORT}
POSTGRES_DB=${POSTGRES_DB}
POSTGRES_USER=${POSTGRES_USER}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
REDIS_HOST=${REDIS_HOST}
REDIS_PORT=${REDIS_PORT}
REDIS_URL=${REDIS_URL}
${REDIS_PASSWORD_LINE}
NATS_URL=${NATS_URL}
APP_ENV=${APP_ENV}
LOG_LEVEL=${LOG_LEVEL}
PORT=${GO_API_PORT}
EOF
run_sudo mv "${WORKDIR}/api.env" "${API_DIR}/.env"
run_sudo chmod 600 "${API_DIR}/.env"

log "Writing worker environment"
cat > "${WORKDIR}/workers.env" <<EOF
POSTGRES_HOST=${POSTGRES_HOST}
POSTGRES_PORT=${POSTGRES_PORT}
POSTGRES_DB=${POSTGRES_DB}
POSTGRES_USER=${POSTGRES_USER}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
NATS_URL=${NATS_URL}
APP_ENV=${APP_ENV}
LOG_LEVEL=${LOG_LEVEL}
SCAN_CHAIN_DB_ID=${SCAN_CHAIN_DB_ID:-1}
SCAN_BATCH_SIZE=${SCAN_BATCH_SIZE:-2000}
SCAN_RPC_URL=${SCAN_RPC_URL:-}
EOF
run_sudo mv "${WORKDIR}/workers.env" "${WORKERS_DIR}/.env"
run_sudo chmod 600 "${WORKERS_DIR}/.env"

log "Deploying Rust node"
run_sudo install -m 755 "${NODE_SRC}/verifiable-rust-chain-node" "${NODE_DIR}/target/release/verifiable-rust-chain-node"

SERVICE_USER="$(id -un)"

log "Refreshing systemd units"
cat > "${WORKDIR}/asset-platform-api.service" <<EOF
[Unit]
Description=BNB Web3 Asset Platform API
After=network.target

[Service]
Type=simple
User=${SERVICE_USER}
WorkingDirectory=${API_DIR}
EnvironmentFile=${API_DIR}/.env
ExecStart=${API_DIR}/api-server-linux-amd64
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

cat > "${WORKDIR}/chain-node.service" <<EOF
[Unit]
Description=BNB Verifiable Rust Chain Node
After=network.target

[Service]
Type=simple
User=${SERVICE_USER}
WorkingDirectory=${NODE_DIR}
ExecStart=${NODE_DIR}/target/release/verifiable-rust-chain-node
Restart=on-failure
RestartSec=5
Environment=RUST_LOG=info
Environment=RPC_PORT=${RUST_RPC_PORT}
Environment=RPC_BIND=0.0.0.0

[Install]
WantedBy=multi-user.target
EOF

cat > "${WORKDIR}/asset-scanner.service" <<EOF
[Unit]
Description=BNB Chain Scanner
After=network.target asset-platform-api.service chain-node.service
Wants=asset-platform-api.service chain-node.service

[Service]
Type=simple
User=${SERVICE_USER}
WorkingDirectory=${WORKERS_DIR}
EnvironmentFile=${WORKERS_DIR}/.env
ExecStart=${WORKERS_DIR}/scanner-linux-amd64
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

cat > "${WORKDIR}/asset-parser.service" <<EOF
[Unit]
Description=BNB Event Parser
After=network.target asset-scanner.service
Wants=asset-scanner.service

[Service]
Type=simple
User=${SERVICE_USER}
WorkingDirectory=${WORKERS_DIR}
EnvironmentFile=${WORKERS_DIR}/.env
ExecStart=${WORKERS_DIR}/parser-linux-amd64
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

cat > "${WORKDIR}/asset-confirm-worker.service" <<EOF
[Unit]
Description=BNB Deposit Confirm Worker
After=network.target asset-parser.service
Wants=asset-parser.service

[Service]
Type=simple
User=${SERVICE_USER}
WorkingDirectory=${WORKERS_DIR}
EnvironmentFile=${WORKERS_DIR}/.env
ExecStart=${WORKERS_DIR}/confirm-worker-linux-amd64
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

cat > "${WORKDIR}/asset-ledger.service" <<EOF
[Unit]
Description=BNB Ledger Service
After=network.target asset-confirm-worker.service
Wants=asset-confirm-worker.service

[Service]
Type=simple
User=${SERVICE_USER}
WorkingDirectory=${WORKERS_DIR}
EnvironmentFile=${WORKERS_DIR}/.env
ExecStart=${WORKERS_DIR}/ledger-service-linux-amd64
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

run_sudo mv "${WORKDIR}/asset-platform-api.service" /etc/systemd/system/asset-platform-api.service
run_sudo mv "${WORKDIR}/chain-node.service" /etc/systemd/system/chain-node.service
run_sudo mv "${WORKDIR}/asset-scanner.service" /etc/systemd/system/asset-scanner.service
run_sudo mv "${WORKDIR}/asset-parser.service" /etc/systemd/system/asset-parser.service
run_sudo mv "${WORKDIR}/asset-confirm-worker.service" /etc/systemd/system/asset-confirm-worker.service
run_sudo mv "${WORKDIR}/asset-ledger.service" /etc/systemd/system/asset-ledger.service

log "Refreshing 1Panel OpenResty config"
cat > "${WORKDIR}/bnb.conf" <<EOF
server {
    listen 80 default_server;
    server_name ${BNB_SERVER_NAME};

    root /www/bnb/frontend;
    index index.html;

    gzip on;
    gzip_types text/plain text/css application/javascript application/json image/svg+xml;
    gzip_min_length 1024;
    gzip_comp_level 5;

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    location /assets/ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    location /api/ {
        proxy_pass http://172.17.0.1:${GO_API_PORT};
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
    }

    location /node/ {
        proxy_pass http://172.17.0.1:${RUST_RPC_PORT}/rpc/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
    }

    location = /health {
        access_log off;
        return 200 "ok\n";
        add_header Content-Type text/plain;
    }
}
EOF
run_sudo mkdir -p "$(dirname "${NGINX_CONF_PATH}")"
run_sudo mv "${WORKDIR}/bnb.conf" "${NGINX_CONF_PATH}"
run_sudo chmod 644 "${NGINX_CONF_PATH}"

OPENRESTY_CONTAINER="$(run_sudo docker ps --format '{{.Names}}' | grep '^1Panel-openresty' | head -n 1 || true)"
if [[ -z "${OPENRESTY_CONTAINER}" ]]; then
    echo "[FATAL] 1Panel OpenResty container not found" >&2
    exit 1
fi

log "Reloading systemd and OpenResty"
run_sudo systemctl daemon-reload
run_sudo systemctl enable asset-platform-api chain-node asset-scanner asset-parser asset-confirm-worker asset-ledger
run_sudo systemctl restart asset-platform-api chain-node asset-scanner asset-parser asset-confirm-worker asset-ledger
run_sudo docker exec "${OPENRESTY_CONTAINER}" /usr/local/openresty/bin/openresty -t
run_sudo docker exec "${OPENRESTY_CONTAINER}" /usr/local/openresty/bin/openresty -s reload

log "Running health checks"
curl -fsS "http://127.0.0.1:${GO_API_PORT}/healthz" >/dev/null
curl -fsS "http://127.0.0.1:${GO_API_PORT}/api/v1/chain-status" >/dev/null
curl -fsS -X POST "http://127.0.0.1:${RUST_RPC_PORT}/rpc/get_block_number" >/dev/null
run_sudo systemctl is-active --quiet asset-scanner
run_sudo systemctl is-active --quiet asset-parser
run_sudo systemctl is-active --quiet asset-confirm-worker
run_sudo systemctl is-active --quiet asset-ledger
run_sudo bash -lc "cd '${WORKERS_DIR}' && set -a && source ./.env && set +a && ./rpc-health-linux-amd64 >/dev/null"

log "Self-hosted deployment completed"
