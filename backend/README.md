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
- Nginx serves the frontend and reverse proxies `/api/*` to the Go app
- frontend and backend share the same public domain
- the server is Debian

Useful deployment files in this folder:

- [deploy/goinvoicer.service](./deploy/goinvoicer.service): example `systemd` unit
- [deploy/goinvoicer.env.example](./deploy/goinvoicer.env.example): example production env file
- [deploy/nginx.conf.example](./deploy/nginx.conf.example): example same-origin Nginx config

### Debian Server Setup

These steps assume:

- Debian 12 or another recent Debian release
- a public domain already pointing at the server
- SSH access to the machine

### 1. Install base packages

```bash
sudo apt update
sudo apt install -y nginx certbot python3-certbot-nginx curl
```

If you want to build the backend manually on the server instead of using GitHub Actions, also install Go:

```bash
sudo apt install -y golang-go
```

### 2. Create the application user and directories

```bash
sudo useradd --system --home /opt/goinvoicer --shell /usr/sbin/nologin goinvoicer || true

sudo mkdir -p /opt/goinvoicer
sudo mkdir -p /etc/goinvoicer
sudo mkdir -p /var/lib/goinvoicer/data
sudo mkdir -p /var/lib/goinvoicer/uploads
sudo mkdir -p /var/www/goinvoicer/releases

sudo chown -R goinvoicer:goinvoicer /opt/goinvoicer
sudo chown -R goinvoicer:goinvoicer /var/lib/goinvoicer
sudo chown -R root:root /etc/goinvoicer
sudo chown -R www-data:www-data /var/www/goinvoicer
```

### 3. Create the production env file

```bash
sudo cp backend/deploy/goinvoicer.env.example /etc/goinvoicer/goinvoicer.env
sudo chmod 600 /etc/goinvoicer/goinvoicer.env
```

Edit `/etc/goinvoicer/goinvoicer.env` and set:

- `PORT=127.0.0.1:4206`
- `APP_BASE_URL=https://your-domain`
- `CORS_ORIGIN=https://your-domain`
- `GOOGLE_REDIRECT_URL=https://your-domain/api/auth/google/callback`
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
sudo cp backend/deploy/nginx.conf.example /etc/nginx/sites-available/goinvoicer.conf
sudo ln -sfn /etc/nginx/sites-available/goinvoicer.conf /etc/nginx/sites-enabled/goinvoicer.conf
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl reload nginx
```

Edit `/etc/nginx/sites-available/goinvoicer.conf` and replace `your-app.example.com` with your real domain.

### 6. Issue the TLS certificate

```bash
sudo certbot --nginx -d your-app.example.com
```

### 7. Put the first backend binary in place

If you are doing the first deployment manually:

```bash
cd backend
go build -o dist/goinvoicer ./cmd
sudo install -m 755 dist/goinvoicer /opt/goinvoicer/goinvoicer
sudo systemctl restart goinvoicer
```

### 8. Put the first frontend build in place

```bash
cd frontend
npm install
npm run build

sudo mkdir -p /var/www/goinvoicer/releases/manual-first
sudo cp -R dist/. /var/www/goinvoicer/releases/manual-first/
sudo ln -sfn /var/www/goinvoicer/releases/manual-first /var/www/goinvoicer/current
sudo systemctl reload nginx
```

### 9. Verify the app

Check that:

- `https://your-domain` serves the frontend
- `https://your-domain/api/auth/me` responds from the Go backend
- Google OAuth callback URL matches your deployed domain
- Stripe webhook URL is set to `https://your-domain/api/billing/stripe/webhook`

### One-Time GitHub Deploy Setup

After the Debian server is ready, configure GitHub so deployments can happen automatically.

### Example Server Directories

- backend binary: `/opt/goinvoicer/goinvoicer`
- backend env file: `/etc/goinvoicer/goinvoicer.env`
- backend working directory: `/var/lib/goinvoicer`
- SQLite database: `/var/lib/goinvoicer/data/goinvoicer.db`
- uploads directory: `/var/lib/goinvoicer/uploads`
- frontend releases: `/var/www/goinvoicer/releases`
- frontend current symlink: `/var/www/goinvoicer/current`

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
- `DEPLOY_BACKEND_DIR` (optional, defaults to `/opt/goinvoicer`)
- `DEPLOY_FRONTEND_ROOT` (optional, defaults to `/var/www/goinvoicer`)
- `DEPLOY_SYSTEMD_SERVICE` (optional, defaults to `goinvoicer`)

Recommended Debian values:

- `DEPLOY_BACKEND_DIR=/opt/goinvoicer`
- `DEPLOY_FRONTEND_ROOT=/var/www/goinvoicer`
- `DEPLOY_SYSTEMD_SERVICE=goinvoicer`

### Repository Secrets

- `DEPLOY_SSH_PRIVATE_KEY`
- `DEPLOY_KNOWN_HOSTS`

`DEPLOY_KNOWN_HOSTS` can be generated with:

```bash
ssh-keyscan -H your-app.example.com
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
- the backend binary should live at `/opt/goinvoicer/goinvoicer`
- the frontend should be served from `/var/www/goinvoicer/current`
- Nginx is already installed and configured
- the SSH user can upload files and restart the backend service

## Recommended Verification

Run the backend gate from the repo root:

```bash
bash scripts/check-backend.sh
```
