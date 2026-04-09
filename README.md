# GoInvoicer

![GoInvoicer Mascot](./goInvoicerMascot.png)

GoInvoicer is a full-stack invoicing app with:

- a Vue 3 frontend in [`frontend`](./frontend)
- a Go + SQLite backend in [`backend`](./backend)
- shared Git hooks and GitHub Actions for local and hosted verification

This README is the quickest path to getting the repo running.

## Deployment Target

The intended production target for this repo is:

- a Debian server
- `systemd` for the Go backend service
- Nginx for HTTPS, static frontend hosting, and reverse proxying `/api/*`
- one deploy root such as `/srv/goinvoicer`
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

## Quick Start

### 1. Clone the repo

```bash
git clone https://github.com/viktorHadz/goInvoice26.git
cd goInvoice26
```

### 2. Install dependencies

```bash
cd frontend
npm install
cd ../backend
go mod download
cd ..
```

### 3. Copy the backend env file

```bash
cp backend/.env.example backend/.env
mkdir -p backend/data
```

### 4. Fill in your local secrets

Edit `backend/.env` and fill in the Google and Stripe values you want to use locally.

Use `backend/.env.example` as the source of truth for the required keys and the default local callback and webhook URLs.

### 5. Start the backend

```bash
cd backend
make run
```

### 6. Start the frontend

Open a second terminal:

```bash
cd frontend
npm run dev
```

The frontend proxies local `/api` requests to the backend during development.

## Local Login And Billing Setup

### Google OAuth

Create a Google OAuth app and register the local callback URL from `backend/.env`.

### Stripe

If you want to test billing locally, point Stripe at the local webhook URL from `backend/.env`.

The Stripe CLI works well for local forwarding:

```bash
stripe listen --forward-to <your local billing webhook URL>
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
- Nginx
- a single deploy root such as `/srv/goinvoicer`
- the backend binary, env file, SQLite data, uploads, and frontend releases kept under that one directory tree

## Where To Read Next

- Backend setup and deployment details: [`backend/README.md`](./backend/README.md)
- Frontend developer guide: [`frontend/README.md`](./frontend/README.md)
