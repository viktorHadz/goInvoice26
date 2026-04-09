#!/usr/bin/env bash

set -euo pipefail

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    cat <<'EOF'
Usage: install-server.sh

Installs the Debian service and Nginx scaffolding for GoInvoicer.

Environment overrides:
  APP_ROOT      default: /srv/goinvoicer
  SERVICE_NAME  default: goinvoicer
  SITE_NAME     default: invoiceandgo.app
  SERVICE_USER  default: goinvoicer
  ENV_SOURCE    default: ./goinvoicer.env.example
  SUDO          default: sudo
EOF
    exit 0
fi

APP_ROOT="${APP_ROOT:-/srv/goinvoicer}"
SERVICE_NAME="${SERVICE_NAME:-goinvoicer}"
SITE_NAME="${SITE_NAME:-invoiceandgo.app}"
SERVICE_USER="${SERVICE_USER:-goinvoicer}"

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ENV_SOURCE="${ENV_SOURCE:-$SCRIPT_DIR/goinvoicer.env.example}"
SUDO="${SUDO:-sudo}"

$SUDO useradd --system --home "$APP_ROOT" --shell /usr/sbin/nologin "$SERVICE_USER" 2>/dev/null || true

$SUDO mkdir -p "$APP_ROOT/data" "$APP_ROOT/uploads" "$APP_ROOT/releases"
$SUDO chown -R "$SERVICE_USER:$SERVICE_USER" "$APP_ROOT"

if [ ! -f "$APP_ROOT/goinvoicer.env" ]; then
    $SUDO install -o "$SERVICE_USER" -g "$SERVICE_USER" -m 600 "$ENV_SOURCE" "$APP_ROOT/goinvoicer.env"
fi

$SUDO install -m 644 "$SCRIPT_DIR/goinvoicer.service" "/etc/systemd/system/$SERVICE_NAME.service"
$SUDO install -m 644 "$SCRIPT_DIR/nginx.conf.example" "/etc/nginx/sites-available/$SITE_NAME"
$SUDO ln -sfn "/etc/nginx/sites-available/$SITE_NAME" "/etc/nginx/sites-enabled/$SITE_NAME"

$SUDO systemctl daemon-reload
$SUDO systemctl enable "$SERVICE_NAME"
$SUDO nginx -t
$SUDO systemctl reload nginx

cat <<EOF
Server scaffolding is in place.

Next steps:
1. Edit $APP_ROOT/goinvoicer.env with the live Google and Stripe values.
2. Build and install the backend binary to $APP_ROOT/goinvoicer.
3. Build the frontend and point $APP_ROOT/current at the release directory.
4. Confirm the Cloudflare Tunnel has hostnames for invoiceandgo.app and www.invoiceandgo.app.
EOF
