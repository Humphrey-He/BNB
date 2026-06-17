#!/usr/bin/env bash
#
# BNB Web3 Career Projects — 1Panel deploy script
#
# 在 SSH 到 101.43.137.178 之后,执行:
#   bash /opt/bnb/deploy.sh
#
# 幂等:任何步骤可重复跑,不会重复创建/覆盖已正确部署的资源。
# 失败退出:任何一步非零退出,脚本立即 abort(不 set -e 的部分会显式处理)。
#
# 前置条件:已经按 PRE_DEPLOY.md 完成 1Panel 面板装 PG/Redis,已经 scp release.tar.gz
#
set -uo pipefail

#==========================================================
# CONFIG(确认 PRE_DEPLOY.md 收集到的实际值)
#==========================================================
PG_HOST="${PG_HOST:-127.0.0.1}"
PG_PORT="${PG_PORT:-5432}"
PG_DB="${PG_DB:-asset_platform}"
PG_USER="${PG_USER:-asset_platform}"
# PG_PASSWORD / REDIS_PASSWORD 不写在这里,运行时 read -s 输入
REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
SERVER_IP="${SERVER_IP:-101.43.127.178}"
NATS_HTTP_PORT="${NATS_HTTP_PORT:-8222}"
NATS_CLIENT_PORT="${NATS_CLIENT_PORT:-4222}"
GO_API_PORT="${GO_API_PORT:-8080}"
RUST_NODE_PORT="${RUST_NODE_PORT:-8081}"

DEPLOY_ROOT="/opt/bnb"
RELEASE_TAR="${DEPLOY_ROOT}/release.tar.gz"
RELEASE_DIR="${DEPLOY_ROOT}/release"
FRONTEND_DIR="${DEPLOY_ROOT}/frontend"
API_DIR="${DEPLOY_ROOT}/api"
NODE_DIR="${DEPLOY_ROOT}/node"
NATS_DATA_DIR="${DEPLOY_ROOT}/nats-data"

#==========================================================
# 提权 wrapper(适配 ubuntu 账号 + root 账号都能跑)
#  - 当前是 root:SUDO=""(直接跑)
#  - 当前是 ubuntu 等普通用户:必须有 sudo 权限
#    (推荐 NOPASSWD,或者愿意每次输入密码)
#==========================================================
SUDO=""
if [[ $EUID -ne 0 ]]; then
    if command -v sudo >/dev/null 2>&1; then
        SUDO="sudo "
        log "检测到非 root 账号,本脚本会通过 sudo 提权执行系统级操作"
    else
        die "当前账号没 sudo,也没 root。请用 1Panel 终端 + root 账号,或给 ubuntu 装 sudo"
    fi
fi

#==========================================================
# 日志
#==========================================================
log()  { printf "\033[1;34m[%s]\033[0m %s\n" "$(date +%H:%M:%S)" "$*"; }
ok()   { printf "\033[1;32m[OK]\033[0m %s\n" "$*"; }
warn() { printf "\033[1;33m[WARN]\033[0m %s\n" "$*"; }
die()  { printf "\033[1;31m[FATAL]\033[0m %s\n" "$*" >&2; exit 1; }

trap 'die "deploy aborted at line $LINENO"' ERR

#==========================================================
# Step 0:环境检查
#==========================================================
log "Step 0/10: 环境检查"
# EUID 检查由上面 SUDO wrapper 处理(非 root 必须有 sudo)
[[ -d "${DEPLOY_ROOT}" ]] || die "目录 ${DEPLOY_ROOT} 不存在,先 sudo mkdir -p ${DEPLOY_ROOT}"
[[ -f "${RELEASE_TAR}" ]] || die "release 包 ${RELEASE_TAR} 不存在,先 scp 上传并 tar -xzf"

command -v docker >/dev/null || die "docker 没装,先在 1Panel 应用商店装 Docker"
command -v systemctl >/dev/null || die "systemctl 不可用,本脚本只支持 systemd 系统"
command -v /usr/local/openresty/bin/openresty >/dev/null || die "OpenResty 没装,在 1Panel 应用商店装"
ok "环境检查通过"

#==========================================================
# Step 1:解压 release 包(幂等)
#==========================================================
log "Step 1/10: 解压 release 包"
${SUDO}mkdir -p "${RELEASE_DIR}"
${SUDO}tar -xzf "${RELEASE_TAR}" -C "${DEPLOY_ROOT}" --skip-old-files
[[ -f "${RELEASE_DIR}/api-server-linux-amd64" ]] || die "release 缺 api-server 二进制"
[[ -d "${RELEASE_DIR}/frontend-dist" ]] || die "release 缺 frontend-dist/"
[[ -d "${RELEASE_DIR}/rust-src" ]] || die "release 缺 rust-src/"
[[ -f "${RELEASE_DIR}/Cargo.toml" ]] || die "release 缺 Cargo.toml"
ok "release 解压完成"

#==========================================================
# Step 2:frontend 静态文件(幂等,rsync 风格)
#==========================================================
log "Step 2/10: 部署 frontend 静态文件"
${SUDO}mkdir -p "${FRONTEND_DIR}"
# 强制覆盖(用 cp -f);不删 frontend 里额外文件以便热重挂
${SUDO}cp -rf "${RELEASE_DIR}/frontend-dist/." "${FRONTEND_DIR}/"
[[ -f "${FRONTEND_DIR}/index.html" ]] || die "frontend index.html 没部署上"
ok "frontend 已部署到 ${FRONTEND_DIR}"

#==========================================================
# Step 3:Go API 部署
#==========================================================
log "Step 3/10: 部署 Go API"
${SUDO}mkdir -p "${API_DIR}"
${SUDO}cp -f "${RELEASE_DIR}/api-server-linux-amd64" "${API_DIR}/"
${SUDO}chmod +x "${API_DIR}/api-server-linux-amd64"
${SUDO}mkdir -p "${API_DIR}/scripts/migrations"
${SUDO}cp -rf "${RELEASE_DIR}/go-scripts/migrations/." "${API_DIR}/scripts/migrations/" 2>/dev/null || warn "go-scripts 缺失,迁移将不会跑"

# 写 .env(密码单独问)
echo -n "请输入 PostgreSQL 密码(输入隐藏): "
read -rs PG_PASSWORD
echo
[[ -n "${PG_PASSWORD}" ]] || die "PG 密码不能为空"

echo -n "请输入 Redis 密码(留空直接回车): "
read -rs REDIS_PASSWORD
echo
# 允许空密码:不强制

if [[ -n "${REDIS_PASSWORD}" ]]; then
    REDIS_URL="redis://:${REDIS_PASSWORD}@${REDIS_HOST}:${REDIS_PORT}"
    REDIS_PASSWORD_LINE="REDIS_PASSWORD=${REDIS_PASSWORD}"
else
    REDIS_URL="redis://${REDIS_HOST}:${REDIS_PORT}"
    REDIS_PASSWORD_LINE="# REDIS_PASSWORD=  # 留空,无密码"
fi

cat > "${API_DIR}/.env" <<EOF
POSTGRES_HOST=${PG_HOST}
POSTGRES_PORT=${PG_PORT}
POSTGRES_DB=${PG_DB}
POSTGRES_USER=${PG_USER}
POSTGRES_PASSWORD=${PG_PASSWORD}
REDIS_HOST=${REDIS_HOST}
REDIS_PORT=${REDIS_PORT}
REDIS_URL=${REDIS_URL}
${REDIS_PASSWORD_LINE}
NATS_URL=nats://127.0.0.1:${NATS_CLIENT_PORT}
APP_ENV=production
LOG_LEVEL=info
PORT=${GO_API_PORT}
EOF
chmod 600 "${API_DIR}/.env"
ok "Go API 部署完成,二进制 + .env 写入 ${API_DIR}"

#==========================================================
# Step 4:跑 PG 迁移(幂等:IF NOT EXISTS 友好)
#==========================================================
log "Step 4/10: 跑 PG 迁移"
if [[ -f "${API_DIR}/scripts/migrations/000001_init_schema_up.sql" ]]; then
    PGPASSWORD="${PG_PASSWORD}" psql -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d "${PG_DB}" \
        -f "${API_DIR}/scripts/migrations/000001_init_schema_up.sql" 2>&1 | tee /tmp/pg-migrate.log
    ok "PG 迁移完成(详细见 /tmp/pg-migrate.log)"
else
    warn "找不到 migration SQL,跳过(API 启动后会自动建表或你手动建)"
fi

#==========================================================
# Step 5:Rust 节点源码 + 编译
#==========================================================
log "Step 5/10: 部署 Rust 节点源码并编译"
${SUDO}mkdir -p "${NODE_DIR}"
${SUDO}cp -rf "${RELEASE_DIR}/rust-src" "${NODE_DIR}/"
${SUDO}cp -f "${RELEASE_DIR}/Cargo.toml" "${NODE_DIR}/"

# 如果是 ubuntu(非 root)跑的,把部署目录 chown 给当前用户,
# 这样后续 cargo build 写 target/ 不需要 sudo
if [[ $EUID -ne 0 ]]; then
    ${SUDO}chown -R "${SUDO_USER:-$(id -un)}:${SUDO_GID:-$(id -gn)}" "${DEPLOY_ROOT}" 2>/dev/null || true
fi

# 装 Rust(如果没装)
if ! command -v cargo >/dev/null 2>&1; then
    log "cargo 不在,装 Rust(约 2 分钟)"
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source "$HOME/.cargo/env"
fi
source "$HOME/.cargo/env" 2>/dev/null || true

# 编译(幂等:cargo build 重复跑没事)
cd "${NODE_DIR}"
cargo build --release 2>&1 | tail -5
[[ -x "${NODE_DIR}/target/release/verifiable-rust-chain-node" ]] \
    || die "Rust 编译失败,看上面 cargo 输出"
ok "Rust 节点编译完成"

#==========================================================
# Step 6:NATS(幂等:已存在则不重建)
#==========================================================
log "Step 6/10: 启 NATS 2.10 (JetStream)"
${SUDO}mkdir -p "${NATS_DATA_DIR}"
if docker ps -a --format '{{.Names}}' | grep -q '^nats$'; then
    log "nats 容器已存在,跳过创建"
    ${SUDO}docker start nats 2>/dev/null || true
else
    ${SUDO}docker run -d \
        --name nats \
        --restart unless-stopped \
        -p 127.0.0.1:${NATS_CLIENT_PORT}:4222 \
        -p 127.0.0.1:${NATS_HTTP_PORT}:8222 \
        -v "${NATS_DATA_DIR}:/data" \
        nats:2.10-alpine \
        -js -sd /data
fi
# 等 NATS 起来
for i in {1..10}; do
    if curl -s "http://127.0.0.1:${NATS_HTTP_PORT}/varz" >/dev/null 2>&1; then
        ok "NATS 已起"
        break
    fi
    sleep 1
done
curl -s "http://127.0.0.1:${NATS_HTTP_PORT}/varz" >/dev/null || die "NATS 起不来"

#==========================================================
# Step 7:systemd 单元(幂等:覆盖写)
#==========================================================
log "Step 7/10: 写 systemd 单元并 enable"

# 服务运行用户:用 SUDO 提权后,如果 SUDO 不为空(非 root),
# 进程以 ubuntu 跑;否则以 root 跑。
SERVICE_USER="$([[ -n "${SUDO}" && "${EUID}" -ne 0 ]] && echo "${SUDO_USER:-ubuntu}" || echo root)"

${SUDO}tee /etc/systemd/system/asset-platform-api.service >/dev/null <<EOF
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

${SUDO}tee /etc/systemd/system/chain-node.service >/dev/null <<EOF
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
Environment=RPC_PORT=${RUST_NODE_PORT}
Environment=RPC_BIND=127.0.0.1

[Install]
WantedBy=multi-user.target
EOF

${SUDO}systemctl daemon-reload
${SUDO}systemctl enable --now asset-platform-api chain-node
sleep 2
ok "systemd 单元已 enable 并启动(进程用户:${SERVICE_USER})"

#==========================================================
# Step 8:OpenResty 反代
#==========================================================
log "Step 8/10: 写 OpenResty 反代 bnb.conf"

NGINX_CONF_DIR="/usr/local/openresty/nginx/conf/conf.d"
[[ -d "${NGINX_CONF_DIR}" ]] || die "OpenResty 配置目录不存在"

# 处理 1Panel 默认站冲突:如果 default.conf 也 listen 80 default_server,移除它的 default_server
if [[ -f "${NGINX_CONF_DIR}/default.conf" ]] && \
   grep -q "default_server" "${NGINX_CONF_DIR}/default.conf"; then
    warn "检测到 default.conf 含 default_server,自动移除关键字避免冲突"
    ${SUDO}sed -i 's/default_server//g' "${NGINX_CONF_DIR}/default.conf"
fi

${SUDO}tee "${NGINX_CONF_DIR}/bnb.conf" >/dev/null <<EOF
server {
    listen 80 default_server;
    server_name ${SERVER_IP} _;

    root ${FRONTEND_DIR};
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
        proxy_pass http://127.0.0.1:${GO_API_PORT}/api/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
    }

    location /node/ {
        proxy_pass http://127.0.0.1:${RUST_NODE_PORT}/rpc;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}
EOF

${SUDO}/usr/local/openresty/bin/openresty -t
${SUDO}/usr/local/openresty/bin/openresty -s reload
ok "OpenResty 已重载"

#==========================================================
# Step 9:放行端口(腾讯云安全组需手动,这里只提示)
#==========================================================
log "Step 9/10: 端口策略提示"
warn "请确认腾讯云安全组入站规则只放行:22(限IP)/80/443"
warn "禁止放行:5432/6379/4222/8080/8081(应保持拒绝)"

#==========================================================
# Step 10:快速自检
#==========================================================
log "Step 10/10: 快速自检"
sleep 2
${SUDO}systemctl is-active --quiet asset-platform-api && ok "asset-platform-api: active" || warn "asset-platform-api 未 active,看 journalctl -u asset-platform-api"
${SUDO}systemctl is-active --quiet chain-node && ok "chain-node: active" || warn "chain-node 未 active,看 journalctl -u chain-node"
docker ps --format '{{.Names}}' | grep -q '^nats$' && ok "nats: running" || warn "nats 未运行"

curl -s "http://127.0.0.1:${NATS_HTTP_PORT}/varz" >/dev/null && ok "NATS monitor OK" || warn "NATS monitor 无响应"
curl -s "http://127.0.0.1:${GO_API_PORT}/healthz" >/dev/null && ok "Go API healthz OK" || warn "Go API healthz 无响应"
curl -s -X POST "http://127.0.0.1:${RUST_NODE_PORT}/rpc" \
    -H "Content-Type: application/json" \
    -d '{"method":"get_mempool_status","params":[],"id":1}' >/dev/null \
    && ok "Rust node RPC OK" || warn "Rust node RPC 无响应"

echo
ok "deploy finished"
echo "下一步:bash /opt/bnb/verify.sh"
echo "前端入口:http://${SERVER_IP}/"
