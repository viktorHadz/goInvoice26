# GoInvoicer

![GoInvoicer Mascot](./goInvoicerMascot.png)

GoInvoicer is a full-stack invoicing app with:

- a Vue 3 frontend in [`frontend`](./frontend)
- a Go + SQLite backend in [`backend`](./backend)
- shared Git hooks and GitHub Actions for local and hosted verification

This README is the quickest path for a developer who wants to run the project locally.

## Deployment Target

The intended production target for this repo is:

- a Debian server
- `systemd` for the Go backend service
- Caddy for HTTPS, static frontend hosting, and reverse proxying `/api/*`
- same-origin hosting on one public domain

## What You Need

- Node.js `22.x` or newer
- npm
- Go `1.25.x`
- a Google OAuth app for sign-in
- Stripe test credentials if you want to exercise billing locally

## Project Layout

- [`frontend`](./frontend): Vue app, local dev server, tests, production build
- [`backend`](./backend): Go API, SQLite database, auth, billing, PDF/DOCX generation
- [`scripts`](./scripts): shared verification commands used by hooks and CI
- [`.githooks`](./.githooks): local `pre-commit` and `pre-push` guards

## Quick Local Setup

### 1. Clone and install dependencies

```bash
git clone https://github.com/viktorHadz/goInvoice26.git
cd goInvoice26

cd frontend
npm install
cd ../backend
go mod download
cd ..
```

### 2. Configure the backend env file

```bash
cd backend
cp .env.example .env
mkdir -p data
```

Edit `backend/.env` and fill in:

- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GOOGLE_REDIRECT_URL`
- `APP_BASE_URL`
- `STRIPE_SECRET_KEY`
- `STRIPE_PUBLISHABLE_KEY`
- `STRIPE_SINGLE_MONTHLY_PRICE_ID`
- `STRIPE_SINGLE_YEARLY_PRICE_ID`
- `STRIPE_TEAM_MONTHLY_PRICE_ID`
- `STRIPE_TEAM_YEARLY_PRICE_ID`
- `STRIPE_TRIAL_DAYS`
- `STRIPE_WEBHOOK_SECRET`

For local development, these values usually look like:

- `APP_BASE_URL=http://localhost:5173`
- `GOOGLE_REDIRECT_URL=http://localhost:4206/api/auth/google/callback`
- `CORS_ORIGIN=http://localhost:5173`

### 3. Start the backend

```bash
cd backend
make run
```

The API listens on `http://localhost:4206`.

### 4. Start the frontend

Open a second terminal:

```bash
cd frontend
npm run dev
```

The app opens on `http://localhost:5173`.

## Local Login And Billing Setup

### Google OAuth

Create a Google OAuth app and add this callback:

```text
http://localhost:4206/api/auth/google/callback
```

### Stripe

If you want to test billing locally, configure a Stripe webhook that points to:

```text
http://localhost:4206/api/billing/stripe/webhook
```

The Stripe CLI works well for local forwarding:

```bash
stripe listen --forward-to localhost:4206/api/billing/stripe/webhook
```

Use `STRIPE_TRIAL_DAYS=7` for the default 7-day trial, or `STRIPE_TRIAL_DAYS=0` if you want checkout to start the paid subscription immediately. Configure `STRIPE_SINGLE_MONTHLY_PRICE_ID` and `STRIPE_SINGLE_YEARLY_PRICE_ID` for the £5/£50 solo prices, plus `STRIPE_TEAM_MONTHLY_PRICE_ID` and `STRIPE_TEAM_YEARLY_PRICE_ID` for the £10/£100 team prices.

## Tests And Verification

Run everything:

```bash
bash scripts/verify-all.sh
```

Run only frontend checks:

```bash
bash scripts/check-frontend.sh
```

Run only backend checks:

```bash
bash scripts/check-backend.sh
```

## Local Git Hooks

This repo includes a `pre-commit` and `pre-push` hook that runs the full verification suite.

Enable them once per clone:

```bash
bash scripts/install-git-hooks.sh
```

After that:

- `git commit` runs the full test gate
- `git push` runs the full test gate again

## CI/CD Overview

Two GitHub Actions workflows are included:

- [`.github/workflows/ci.yml`](./.github/workflows/ci.yml): runs frontend and backend checks on every push and pull request
- [`.github/workflows/deploy.yml`](./.github/workflows/deploy.yml): deploys to your server after `CI` passes on `main`, or manually through `workflow_dispatch`

The deploy workflow is designed around a Debian server with:

- `systemd`
- Caddy
- a writable backend directory such as `/opt/goinvoicer`
- a frontend web root such as `/var/www/goinvoicer/current`

## Where To Read Next

- Backend setup and deployment details: [`backend/README.md`](./backend/README.md)
- Frontend developer guide: [`frontend/README.md`](./frontend/README.md)
