#!/usr/bin/env bash
#
# BNB 部署后验证脚本
# 跑完输出 6 项检查,绿/红可视化
#
# 用法:bash /opt/bnb/verify.sh
#
set -uo pipefail

GO_API_PORT="${GO_API_PORT:-8080}"
RUST_NODE_PORT="${RUST_NODE_PORT:-8081}"
NATS_HTTP_PORT="${NATS_HTTP_PORT:-8222}"
SERVER_IP="${SERVER_IP:-101.43.127.178}"

# 提权 wrapper(同 deploy.sh)
SUDO=""
if [[ $EUID -ne 0 ]] && command -v sudo >/dev/null 2>&1; then
    SUDO="sudo "
fi

PASS=0
FAIL=0

green() { printf "\033[1;32m%s\033[0m" "$1"; }
red()   { printf "\033[1;31m%s\033[0m" "$1"; }
yellow(){ printf "\033[1;33m%s\033[0m" "$1"; }

check() {
    local name="$1"
    local cmd="$2"
    if eval "${cmd}" >/dev/null 2>&1; then
        echo "  $(green '✓') ${name}"
        PASS=$((PASS+1))
    else
        echo "  $(red '✗') ${name}"
        FAIL=$((FAIL+1))
    fi
}

hint() { echo "  $(yellow '→') $1"; }

echo "================================================"
echo " BNB deploy verify @ $(date '+%F %T')"
echo " server: ${SERVER_IP}"
echo "================================================"
echo

echo "[1/6] systemd 单元"
check "asset-platform-api active" "${SUDO}systemctl is-active --quiet asset-platform-api"
check "chain-node active" "${SUDO}systemctl is-active --quiet chain-node"
[[ $FAIL -gt 0 ]] && hint "看日志:journalctl -u asset-platform-api -n 30 --no-pager"
echo

echo "[2/6] Docker 容器"
check "nats 容器 running" "docker ps --format '{{.Names}}' | grep -q '^nats$'"
echo

echo "[3/6] 进程端口"
check "Go API 监听 127.0.0.1:${GO_API_PORT}" "${SUDO}ss -lntp 2>/dev/null | grep -q '127.0.0.1:${GO_API_PORT}'"
check "Rust node 监听 127.0.0.1:${RUST_NODE_PORT}" "${SUDO}ss -lntp 2>/dev/null | grep -q '127.0.0.1:${RUST_NODE_PORT}'"
check "NATS 监听 127.0.0.1:${NATS_HTTP_PORT}" "${SUDO}ss -lntp 2>/dev/null | grep -q '127.0.0.1:${NATS_HTTP_PORT}'"
echo

echo "[4/6] 服务健康"
check "Go API healthz" "curl -sf http://127.0.0.1:${GO_API_PORT}/healthz"
check "Go API chain-status" "curl -sf http://127.0.0.1:${GO_API_PORT}/api/v1/chain-status"
check "Rust node mempool_status" "curl -sf -X POST http://127.0.0.1:${RUST_NODE_PORT}/rpc -H 'Content-Type: application/json' -d '{\"method\":\"get_mempool_status\",\"params\":[],\"id\":1}'"
check "NATS varz" "curl -sf http://127.0.0.1:${NATS_HTTP_PORT}/varz"

# 尝试用 .env 里的 Redis 密码测一下连通性(不强制)
if [[ -f /opt/bnb/api/.env ]] && grep -q "^REDIS_PASSWORD=" /opt/bnb/api/.env; then
    REDIS_PW=$(grep "^REDIS_PASSWORD=" /opt/bnb/api/.env | cut -d= -f2-)
    if command -v redis-cli >/dev/null 2>&1; then
        if [[ -n "${REDIS_PW}" ]]; then
            check "Redis ping (带密码)" "redis-cli -h 127.0.0.1 -p 6379 -a \"\${REDIS_PW}\" ping 2>/dev/null | grep -q PONG"
        else
            check "Redis ping (无密码)" "redis-cli -h 127.0.0.1 -p 6379 ping 2>/dev/null | grep -q PONG"
        fi
    else
        log "redis-cli 不在,跳过 Redis ping"
    fi
fi
echo

echo "[5/6] 反代可达"
check "OpenResty frontend /" "curl -sf http://127.0.0.1/ -o /dev/null"
check "OpenResty 反代 /api/v1/chain-status" "curl -sf http://127.0.0.1/api/v1/chain-status"
check "OpenResty 反代 /node" "curl -sf -X POST http://127.0.0.1/node -H 'Content-Type: application/json' -d '{\"method\":\"get_mempool_status\",\"params\":[],\"id\":1}'"
echo

echo "[6/6] 安全"
HAVE_5432=$(ss -lntp 2>/dev/null | grep -c '0.0.0.0:5432')
HAVE_6379=$(ss -lntp 2>/dev/null | grep -c '0.0.0.0:6379')
HAVE_4222=$(ss -lntp 2>/dev/null | grep -c '0.0.0.0:4222')
HAVE_8080=$(ss -lntp 2>/dev/null | grep -c '0.0.0.0:8080')
HAVE_8081=$(ss -lntp 2>/dev/null | grep -c '0.0.0.0:8081')
if [[ $HAVE_5432 -eq 0 && $HAVE_6379 -eq 0 && $HAVE_4222 -eq 0 && $HAVE_8080 -eq 0 && $HAVE_8081 -eq 0 ]]; then
    echo "  $(green '✓') 5432/6379/4222/8080/8081 都未暴露 0.0.0.0"
    PASS=$((PASS+1))
else
    echo "  $(red '✗') 以下端口暴露 0.0.0.0,需立即在腾讯云安全组/PG/Redis 配置里加 listen_addresses"
    [[ $HAVE_5432 -gt 0 ]] && hint "PG 5432 暴露"
    [[ $HAVE_6379 -gt 0 ]] && hint "Redis 6379 暴露"
    [[ $HAVE_4222 -gt 0 ]] && hint "NATS 4222 暴露"
    [[ $HAVE_8080 -gt 0 ]] && hint "Go API 8080 暴露"
    [[ $HAVE_8081 -gt 0 ]] && hint "Rust 8081 暴露"
    FAIL=$((FAIL+1))
fi
echo

echo "================================================"
echo " 结果: $(green "${PASS} 通过") / $([[ $FAIL -eq 0 ]] && green "${FAIL} 失败" || red "${FAIL} 失败")"
echo "================================================"

if [[ $FAIL -gt 0 ]]; then
    echo
    yellow "修复建议:"
    echo "  1. 跑 bash /opt/bnb/deploy.sh 重试(幂等)"
    echo "  2. 看具体服务的 journalctl:journalctl -u <unit> -n 50 --no-pager"
    echo "  3. 重置某个服务:systemctl restart <unit>"
    echo "  4. 全部回滚:bash /opt/bnb/rollback.sh"
    exit 1
fi

echo
green "全部通过。打开 http://${SERVER_IP}/ 即可访问。"
exit 0
