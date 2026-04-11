#!/usr/bin/env bash

set -euo pipefail

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    cat <<'EOF'
Usage: bootstrap-invoiceandgo.sh

Builds the frontend and backend, renders a production env from backend/.env,
installs the service and Nginx scaffolding, backs up the live database when present,
and publishes a release to /srv/goinvoicer.

Environment overrides:
  APP_URL       default: invoiceandgo.app
  APP_ROOT      default: /srv/goinvoicer
  SERVICE_NAME  default: goinvoicer
  SERVICE_USER  default: goinvoicer
  SITE_NAME     default: invoiceandgo.app
  SOURCE_ENV    default: <repo>/backend/.env
  NODE_BIN_DIR  default: /home/vik-usr/.nvm/versions/node/v22.13.1/bin
  SUDO          default: sudo
EOF
    exit 0
fi

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../.." && pwd)"

APP_URL="${APP_URL:-invoiceandgo.app}"
APP_ROOT="${APP_ROOT:-/srv/goinvoicer}"
SERVICE_NAME="${SERVICE_NAME:-goinvoicer}"
SERVICE_USER="${SERVICE_USER:-goinvoicer}"
SITE_NAME="${SITE_NAME:-$APP_URL}"
SOURCE_ENV="${SOURCE_ENV:-$REPO_ROOT/backend/.env}"
SUDO="${SUDO:-sudo}"

NODE_BIN_DIR="${NODE_BIN_DIR:-/home/vik-usr/.nvm/versions/node/v22.13.1/bin}"
if [ -d "$NODE_BIN_DIR" ]; then
    export PATH="$NODE_BIN_DIR:$PATH"
fi
export PATH="/usr/local/go/bin:$PATH"

if [ ! -f "$SOURCE_ENV" ]; then
    echo "Source env file not found: $SOURCE_ENV" >&2
    exit 1
fi

command -v npm >/dev/null 2>&1 || {
    echo "npm is required on PATH before running this script." >&2
    exit 1
}
command -v go >/dev/null 2>&1 || {
    echo "go is required on PATH before running this script." >&2
    exit 1
}

PROD_ENV="$(mktemp /tmp/goinvoicer.env.XXXXXX)"
BACKEND_BIN="$(mktemp /tmp/goinvoicer.bin.XXXXXX)"
RELEASE_ID="manual-$(date +%Y%m%d-%H%M%S)"

cleanup() {
    rm -f "$PROD_ENV" "$BACKEND_BIN"
}
trap cleanup EXIT

cp "$SOURCE_ENV" "$PROD_ENV"
sed -i \
  -e 's|^ENV=.*|ENV=production|' \
  -e 's|^PORT=.*|PORT=127.0.0.1:4206|' \
  -e "s|^DB_PATH=.*|DB_PATH=$APP_ROOT/data/goinvoicer.db|" \
  -e "s|^CORS_ORIGIN=.*|CORS_ORIGIN=https://$APP_URL|" \
  -e "s|^APP_BASE_URL=.*|APP_BASE_URL=https://$APP_URL|" \
  -e "s|^GOOGLE_REDIRECT_URL=.*|GOOGLE_REDIRECT_URL=https://$APP_URL/api/auth/google/callback|" \
  "$PROD_ENV"

(
    cd "$REPO_ROOT/frontend"
    npm run build
)

(
    cd "$REPO_ROOT/backend"
    go build -o "$BACKEND_BIN" ./cmd
)

ENV_SOURCE="$PROD_ENV" \
APP_ROOT="$APP_ROOT" \
SERVICE_NAME="$SERVICE_NAME" \
SERVICE_USER="$SERVICE_USER" \
SITE_NAME="$SITE_NAME" \
SUDO="$SUDO" \
    "$SCRIPT_DIR/install-server.sh"

LIVE_ENV_PATH="$APP_ROOT/goinvoicer.env"
LIVE_DB_PATH="$($SUDO sh -c "sed -n 's/^DB_PATH=//p' '$LIVE_ENV_PATH' | head -n1" 2>/dev/null || true)"
if [ -z "$LIVE_DB_PATH" ]; then
    LIVE_DB_PATH="$(sed -n 's/^DB_PATH=//p' "$PROD_ENV" | head -n1)"
fi
if [ -n "$LIVE_DB_PATH" ] && $SUDO test -f "$LIVE_DB_PATH"; then
    DB_BACKUP_PATH="${LIVE_DB_PATH}.bak.$(date +%Y%m%d-%H%M%S)"
    echo "Backing up live database to $DB_BACKUP_PATH"
    $SUDO cp "$LIVE_DB_PATH" "$DB_BACKUP_PATH"
fi

$SUDO install -m 755 "$BACKEND_BIN" "$APP_ROOT/goinvoicer"
$SUDO install -d -m 755 "$APP_ROOT/releases/$RELEASE_ID"
$SUDO cp -R "$REPO_ROOT/frontend/dist/." "$APP_ROOT/releases/$RELEASE_ID/"
$SUDO ln -sfn "$APP_ROOT/releases/$RELEASE_ID" "$APP_ROOT/current"
$SUDO systemctl restart "$SERVICE_NAME"
$SUDO systemctl --no-pager --full status "$SERVICE_NAME" --lines=20
$SUDO nginx -t
$SUDO systemctl reload nginx

cat <<EOF
Deployment finished.

Verify:
  https://$APP_URL
  https://$APP_URL/api/auth/me

Cloudflare Tunnel still needs public hostnames that forward to:
  http://127.0.0.1:80
EOF
