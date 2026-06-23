#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
UNITS_DIR="${REPO_ROOT}/ops/systemd"

require_root() {
    if [[ "${EUID}" -ne 0 ]]; then
        echo "[FATAL] please run as root" >&2
        exit 1
    fi
}

install_unit() {
    local src="$1"
    local name
    name="$(basename "${src}")"
    install -m 644 "${src}" "/etc/systemd/system/${name}"
    echo "[OK] installed ${name}"
}

require_root

for unit in \
    asset-platform-api.service \
    asset-scanner.service \
    asset-parser.service \
    asset-confirm-worker.service \
    asset-ledger.service \
    asset-withdrawal-worker.service \
    asset-broadcaster.service \
    chain-node.service
do
    install_unit "${UNITS_DIR}/${unit}"
done

systemctl daemon-reload

cat <<'EOF'
[OK] systemd templates installed.

Next:
1. Copy env files into /home/ubuntu/opt/bnb/{api,workers,node}/.env
2. Ensure binaries exist in /home/ubuntu/opt/bnb/{api,workers,node}
3. Enable services, for example:
   systemctl enable --now asset-platform-api asset-scanner asset-parser asset-confirm-worker asset-ledger asset-withdrawal-worker asset-broadcaster chain-node
EOF
