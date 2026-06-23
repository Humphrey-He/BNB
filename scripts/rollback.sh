#!/usr/bin/env bash
#
# BNB 部署回滚脚本
# 作用:停所有 systemd 单元、删 NATS 容器、删 OpenResty 反代、保留 PG/Redis 数据
# 不删:PG 数据(1Panel 管)、Redis 数据(1Panel 管)、/opt/bnb/nats-data(可选删)
#
# 用法:bash /opt/bnb/rollback.sh [--purge-data]
#   --purge-data:顺带删 nats-data、frontend 静态、api/node 二进制
#                 不删 PG/Redis 数据(1Panel 自己管)
#
set -uo pipefail

log()  { printf "\033[1;34m[%s]\033[0m %s\n" "$(date +%H:%M:%S)" "$*"; }
ok()   { printf "\033[1;32m[OK]\033[0m %s\n" "$*"; }
warn() { printf "\033[1;33m[WARN]\033[0m %s\n" "$*"; }
die()  { printf "\033[1;31m[FATAL]\033[0m %s\n" "$*" >&2; exit 1; }

[[ $EUID -eq 0 ]] || die "请用 root 跑"

PURGE_DATA=false
if [[ "${1:-}" == "--purge-data" ]]; then
    PURGE_DATA=true
    warn "将额外清理 /opt/bnb/{frontend,api,node,nats-data}(PG/Redis 数据不删)"
    echo -n "确认?(yes/no): "
    read -r ans
    [[ "$ans" == "yes" ]] || die "取消"
fi

DEPLOY_ROOT="/opt/bnb"

#==========================================================
# 1.停 systemd 单元
#==========================================================
log "1. 停 systemd 单元"
for svc in \
    asset-platform-api \
    asset-scanner \
    asset-parser \
    asset-confirm-worker \
    asset-ledger \
    asset-withdrawal-worker \
    asset-broadcaster \
    chain-node
do
    if systemctl list-unit-files | grep -q "^${svc}.service"; then
        systemctl stop "${svc}" 2>/dev/null || warn "${svc} stop 失败(可能没在跑)"
        systemctl disable "${svc}" 2>/dev/null || warn "${svc} disable 失败"
        rm -f "/etc/systemd/system/${svc}.service"
        ok "${svc} 已 stop + disable + 删单元"
    else
        log "${svc}.service 不存在,跳过"
    fi
done
systemctl daemon-reload
systemctl reset-failed 2>/dev/null || true

#==========================================================
# 2.停 NATS 容器
#==========================================================
log "2. 停 NATS 容器"
if docker ps -a --format '{{.Names}}' | grep -q '^nats$'; then
    docker stop nats
    docker rm nats
    ok "nats 容器已删"
else
    log "nats 容器不存在,跳过"
fi

#==========================================================
# 3.删 OpenResty 反代
#==========================================================
log "3. 删 OpenResty 反代 bnb.conf"
NGINX_CONF="/usr/local/openresty/nginx/conf/conf.d/bnb.conf"
if [[ -f "$NGINX_CONF" ]]; then
    rm -f "$NGINX_CONF"
    # 恢复 1Panel default.conf 的 default_server(如果之前被 sed 改过)
    DEFAULT_CONF="/usr/local/openresty/nginx/conf/conf.d/default.conf"
    if [[ -f "$DEFAULT_CONF" ]] && ! grep -q "default_server" "$DEFAULT_CONF"; then
        # 检查 listen 80 行
        if grep -q "listen\s*80" "$DEFAULT_CONF"; then
            warn "恢复 default.conf 的 default_server 关键字"
            sed -i 's/listen 80;/listen 80 default_server;/' "$DEFAULT_CONF" 2>/dev/null || true
        fi
    fi
    /usr/local/openresty/bin/openresty -t && /usr/local/openresty/bin/openresty -s reload
    ok "OpenResty 已 reload,bnb.conf 已删"
else
    log "bnb.conf 不存在,跳过"
fi

#==========================================================
# 4.(可选)清数据
#==========================================================
if $PURGE_DATA; then
    log "4. 清部署目录"
    rm -rf "${DEPLOY_ROOT}/frontend"
    rm -rf "${DEPLOY_ROOT}/api"
    rm -rf "${DEPLOY_ROOT}/node"
    rm -rf "${DEPLOY_ROOT}/nats-data"
    rm -rf "${DEPLOY_ROOT}/release"
    rm -f "${DEPLOY_ROOT}/release.tar.gz"
    ok "/opt/bnb/{frontend,api,node,nats-data,release,release.tar.gz} 已删"
else
    log "4. 保留 /opt/bnb/{frontend,api,node,nats-data,release},要彻底清请加 --purge-data"
fi

#==========================================================
# 5.提示 PG/Redis 是 1Panel 管的
#==========================================================
echo
warn "PG/Redis 是在 1Panel 应用商店装的,本脚本不动。"
warn "要彻底清 PG 数据:1Panel → 数据库 → asset_platform → 删除(慎!)"
warn "要彻底清 Redis 数据:1Panel → 应用商店 → Redis → 卸载(慎!)"

echo
ok "rollback finished"
echo
echo "验证残留:"
ss -lntp 2>/dev/null | grep -E ":(8080|8081|4222|8222)\b" || echo "  无残留端口 ✓"
docker ps -a --format '{{.Names}}' | grep -q '^nats$' && echo "  nats 还在!" || echo "  nats 已清 ✓"
ls /etc/systemd/system/asset-platform-api.service 2>/dev/null && echo "  api 单元还在!" || echo "  api 单元已清 ✓"
ls /etc/systemd/system/asset-broadcaster.service 2>/dev/null && echo "  broadcaster 单元还在!" || echo "  broadcaster 单元已清 ✓"
ls /etc/systemd/system/chain-node.service 2>/dev/null && echo "  node 单元还在!" || echo "  node 单元已清 ✓"
