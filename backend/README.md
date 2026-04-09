# GoInvoicer Backend

![GoInvoicer Mascot](./goInvoicerMascot.png)

This is the Go API for GoInvoicer.

It is responsible for:

- Google sign-in and session handling
- workspace, team, settings, and billing logic
- invoice creation, revisions, payments, and exports
- SQLite persistence
- PDF and DOCX generation

## Requirements

- Go `1.25.x`
- SQLite support through `github.com/mattn/go-sqlite3`
- Google OAuth credentials
- Stripe credentials if billing is enabled

## Local Setup

### 1. Create the env file

```bash
cp .env.example .env
mkdir -p data
```

### 2. Fill in the required values

The backend reads from `.env` in development.

Important values:

- `DB_PATH`
- `APP_BASE_URL`
- `CORS_ORIGIN`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GOOGLE_REDIRECT_URL`
- `STRIPE_PUBLISHABLE_KEY`
- `STRIPE_SECRET_KEY`
- `STRIPE_SINGLE_MONTHLY_PRICE_ID`
- `STRIPE_SINGLE_YEARLY_PRICE_ID`
- `STRIPE_TEAM_MONTHLY_PRICE_ID`
- `STRIPE_TEAM_YEARLY_PRICE_ID`
- `STRIPE_TRIAL_DAYS`
- `STRIPE_WEBHOOK_SECRET`

Typical local values:

```text
APP_BASE_URL=http://localhost:5173
CORS_ORIGIN=http://localhost:5173
GOOGLE_REDIRECT_URL=http://localhost:4206/api/auth/google/callback
DB_PATH=./data/goinvoicer.db
```

### 3. Run the API

```bash
make run
```

The API listens on `http://localhost:4206`.

## Useful Commands

Run the API:

```bash
make run
```

Build the production binary:

```bash
make build
```

Run backend tests:

```bash
make test
```

Run static checks:

```bash
make vet
```

## Auth And Billing Notes

### Google OAuth

Add this callback URL to your Google OAuth app for local development:

```text
http://localhost:4206/api/auth/google/callback
```

### Stripe

For local billing tests, point Stripe webhooks to:

```text
http://localhost:4206/api/billing/stripe/webhook
```

If you use the Stripe CLI:

```bash
stripe listen --forward-to localhost:4206/api/billing/stripe/webhook
```

Use `STRIPE_TRIAL_DAYS=7` for the default 7-day trial, or `STRIPE_TRIAL_DAYS=0` if you want checkout to start the paid subscription immediately. Configure the four Stripe price IDs for single monthly, single yearly, team monthly, and team yearly billing.

## Production Deployment

The intended production setup is:

- the Go API runs as a `systemd` service
- Nginx listens on `127.0.0.1:80`, serves the frontend, and reverse proxies `/api/*` to the Go app
- Cloudflare Tunnel publishes the public HTTPS hostname and forwards to local Nginx
- frontend and backend share the same public domain
- the server is Debian

The default production layout is intentionally simple and keeps everything under one directory:

- `/srv/goinvoicer/goinvoicer`
- `/srv/goinvoicer/goinvoicer.env`
- `/srv/goinvoicer/data/goinvoicer.db`
- `/srv/goinvoicer/uploads`
- `/srv/goinvoicer/releases`
- `/srv/goinvoicer/current`

Useful deployment files in this folder:

- [deploy/goinvoicer.service](./deploy/goinvoicer.service): example `systemd` unit
- [deploy/goinvoicer.env.example](./deploy/goinvoicer.env.example): example production env file
- [deploy/nginx.conf.example](./deploy/nginx.conf.example): example same-origin Nginx config
- [deploy/install-server.sh](./deploy/install-server.sh): installs the service and Nginx scaffolding on a Debian box
- [deploy/bootstrap-invoiceandgo.sh](./deploy/bootstrap-invoiceandgo.sh): builds the app, renders a production env from `backend/.env`, and installs it

### Debian Server Setup

These steps assume:

- Debian 12 or another recent Debian release
- a public hostname of `invoiceandgo.app`
- a Cloudflare Tunnel already installed on the box
- the tunnel forwards `invoiceandgo.app` to `http://127.0.0.1:80`
- SSH access to the machine

### 1. Install base packages

```bash
sudo apt update
sudo apt install -y nginx curl
```

If you want to build the backend manually on the server instead of using GitHub Actions, also install Go:

```bash
sudo apt install -y golang-go
```

If this server already has the repo checked out, Go installed, Node installed, Nginx installed, and a working sudo user, you can use the one-shot bootstrap script instead of running the rest of the steps by hand:

```bash
cd backend
./deploy/bootstrap-invoiceandgo.sh
```

### 2. Create the application user and directories

```bash
sudo useradd --system --home /srv/goinvoicer --shell /usr/sbin/nologin goinvoicer || true

sudo mkdir -p /srv/goinvoicer/certs
sudo mkdir -p /srv/goinvoicer/data
sudo mkdir -p /srv/goinvoicer/uploads
sudo mkdir -p /srv/goinvoicer/releases

sudo chown -R goinvoicer:goinvoicer /srv/goinvoicer
```

### 3. Create the production env file

```bash
sudo cp backend/deploy/goinvoicer.env.example /srv/goinvoicer/goinvoicer.env
sudo chown goinvoicer:goinvoicer /srv/goinvoicer/goinvoicer.env
sudo chmod 600 /srv/goinvoicer/goinvoicer.env
```

Edit `/srv/goinvoicer/goinvoicer.env` and set:

- `PORT=127.0.0.1:4206`
- `APP_BASE_URL=https://invoiceandgo.app`
- `CORS_ORIGIN=https://invoiceandgo.app`
- `GOOGLE_REDIRECT_URL=https://invoiceandgo.app/api/auth/google/callback`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `STRIPE_PUBLISHABLE_KEY`
- `STRIPE_SECRET_KEY`
- `STRIPE_SINGLE_MONTHLY_PRICE_ID`
- `STRIPE_SINGLE_YEARLY_PRICE_ID`
- `STRIPE_TEAM_MONTHLY_PRICE_ID`
- `STRIPE_TEAM_YEARLY_PRICE_ID`
- `STRIPE_TRIAL_DAYS`
- `STRIPE_WEBHOOK_SECRET`

### 4. Install the backend service

```bash
sudo cp backend/deploy/goinvoicer.service /etc/systemd/system/goinvoicer.service
sudo systemctl daemon-reload
sudo systemctl enable goinvoicer
```

### 5. Configure Nginx

```bash
sudo cp backend/deploy/nginx.conf.example /etc/nginx/sites-available/invoiceandgo.app
sudo ln -sfn /etc/nginx/sites-available/invoiceandgo.app /etc/nginx/sites-enabled/invoiceandgo.app
sudo nginx -t
sudo systemctl reload nginx
```

The example file is already written for the existing tunnel-backed Nginx pattern on this server:

- Nginx listens only on `127.0.0.1:80`
- `www.invoiceandgo.app` redirects to `invoiceandgo.app`
- the frontend is served from `/srv/goinvoicer/current`
- `/api/*` is proxied to `127.0.0.1:4206`
- `CF-Connecting-IP` is trusted from the local tunnel process
### 6. Confirm the Cloudflare Tunnel hostname

This machine is already using a token-managed `cloudflared.service`, so there may be no local `/etc/cloudflared/config.yml` to edit.

In the Cloudflare Zero Trust dashboard, confirm the tunnel has these public hostnames:

- `invoiceandgo.app` -> `http://127.0.0.1:80`
- `www.invoiceandgo.app` -> `http://127.0.0.1:80`

If the tunnel is already forwarding all hostnames to local Nginx, you only need to make sure the DNS hostname is attached to the tunnel.

### 7. Match the Cloudflare SSL mode

In Cloudflare, keep the hostname proxied and use `Full` or `Full (strict)` at the edge. Nginx itself stays on loopback HTTP because Cloudflare Tunnel terminates the public HTTPS connection.

### 8. Put the first backend binary in place

If you are doing the first deployment manually:

```bash
cd backend
go build -o dist/goinvoicer ./cmd
sudo install -m 755 dist/goinvoicer /srv/goinvoicer/goinvoicer
sudo systemctl restart goinvoicer
```

### 9. Put the first frontend build in place

```bash
cd frontend
npm install
npm run build

sudo mkdir -p /srv/goinvoicer/releases/manual-first
sudo cp -R dist/. /srv/goinvoicer/releases/manual-first/
sudo ln -sfn /srv/goinvoicer/releases/manual-first /srv/goinvoicer/current
sudo systemctl reload nginx
```

### 10. Verify the app

Check that:

- `https://invoiceandgo.app` serves the frontend
- `https://invoiceandgo.app/api/auth/me` responds from the Go backend
- Cloudflare Tunnel is forwarding the hostname to local Nginx
- Google OAuth callback URL matches `https://invoiceandgo.app/api/auth/google/callback`
- Stripe webhook URL is set to `https://invoiceandgo.app/api/billing/stripe/webhook`

### One-Time GitHub Deploy Setup

After the Debian server is ready, configure GitHub so deployments can happen automatically.

### Example Server Directories

- deploy root: `/srv/goinvoicer`
- backend binary: `/srv/goinvoicer/goinvoicer`
- backend env file: `/srv/goinvoicer/goinvoicer.env`
- SQLite database: `/srv/goinvoicer/data/goinvoicer.db`
- uploads directory: `/srv/goinvoicer/uploads`
- frontend releases: `/srv/goinvoicer/releases`
- frontend current symlink: `/srv/goinvoicer/current`

## GitHub Deploy Workflow

This repo includes [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml).

It:

1. waits for the `CI` workflow to pass on `main`
2. builds the frontend and backend
3. uploads the artifacts to your server over SSH
4. updates the frontend `current` symlink
5. restarts the backend `systemd` service

To use it, configure these GitHub repository settings.

### Repository Variables

- `DEPLOY_HOST`
- `DEPLOY_PORT` (optional, defaults to `22`)
- `DEPLOY_USER` (optional, defaults to `root`)
- `DEPLOY_BACKEND_DIR` (optional, defaults to `/srv/goinvoicer`)
- `DEPLOY_FRONTEND_ROOT` (optional, defaults to `/srv/goinvoicer`)
- `DEPLOY_SYSTEMD_SERVICE` (optional, defaults to `goinvoicer`)

Recommended Debian values:

- `DEPLOY_BACKEND_DIR=/srv/goinvoicer`
- `DEPLOY_FRONTEND_ROOT=/srv/goinvoicer`
- `DEPLOY_SYSTEMD_SERVICE=goinvoicer`

### Repository Secrets

- `DEPLOY_SSH_PRIVATE_KEY`
- `DEPLOY_KNOWN_HOSTS`

`DEPLOY_KNOWN_HOSTS` can be generated with:

```bash
ssh-keyscan -H invoiceandgo.app
```

### Important Permission Note

The SSH user used by the deploy workflow must be able to:

- write to the backend and frontend deploy directories
- restart the backend service with `systemctl`

The simplest approach is either:

- use a trusted deploy user with passwordless `sudo` for the service restart, or
- use root over SSH if that matches your server policy

For a Debian self-hosted box, starting with `root` as the deploy user is the simplest path. You can harden that later once the deployment flow is stable.

## Deploying Through GitHub Actions

Once the server and GitHub secrets are ready:

1. push to `main`
2. wait for `CI` to pass
3. the `Deploy` workflow will start automatically
4. check the Actions tab to confirm the rollout succeeded

You can also trigger the deploy workflow manually from the GitHub Actions UI.

## What The Deploy Workflow Expects

The current deploy workflow assumes:

- Debian or another Linux distro with `systemd`
- the backend binary should live at `/srv/goinvoicer/goinvoicer`
- the frontend should be served from `/srv/goinvoicer/current`
- Nginx is already installed and configured
- the SSH user can upload files and restart the backend service

## Recommended Verification

Run the backend gate from the repo root:

```bash
bash scripts/check-backend.sh
```
