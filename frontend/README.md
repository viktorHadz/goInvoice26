# GoInvoicer Frontend

This is the Vue 3 frontend for GoInvoicer.

It handles:

- login and signup flows
- clients, invoices, revisions, and exports
- billing UI
- team and settings screens

## Requirements

- Node.js `22.x` or newer
- npm

## Install

```bash
npm install
```

## Run Locally

```bash
npm run dev
```

The frontend runs at `http://localhost:5173`.

In local development, Vite proxies `/api` requests to the Go backend on `http://localhost:4206`.

## Available Commands

Start the local dev server:

```bash
npm run dev
```

Build the production frontend:

```bash
npm run build
```

Preview the production build locally:

```bash
npm run preview
```

Run type-checking:

```bash
npm run type-check
```

Run the frontend test suite:

```bash
npm test
```

Run linting:

```bash
npm run lint
```

## Frontend Environment Notes

At the moment, the frontend does not depend on a required production `VITE_*` config set for API access.

That is intentional:

- local dev uses the Vite `/api` proxy
- production uses same-origin hosting, where the frontend and backend sit behind the same domain

There is a local `frontend/.env` file in this repo, but it is not part of the required production setup.

## Production Hosting Model

The production frontend is built into `frontend/dist`.

For deployment, the intended setup is:

1. build the frontend with `npm run build`
2. serve `frontend/dist` with Nginx
3. reverse proxy `/api/*` to the Go backend
4. use SPA fallback so app routes such as `/app/invoice` still resolve to `index.html`

An Nginx example lives at [backend/deploy/nginx.conf.example](../backend/deploy/nginx.conf.example).

## Debian Production Notes

The production setup for this repo assumes a Debian server.

On Debian, the frontend is expected to end up here:

```text
/srv/goinvoicer/current
```

Nginx then serves that directory and forwards `/api/*` to the backend on `127.0.0.1:4206`.

The GitHub deploy workflow updates the frontend by:

1. building `frontend/dist` in GitHub Actions
2. uploading it to the Debian server
3. extracting it into a release folder
4. repointing `/srv/goinvoicer/current` to the new release

## Recommended Developer Workflow

From the repo root, run the shared frontend checks with:

```bash
bash scripts/check-frontend.sh
```

That matches what local Git hooks and GitHub Actions run.
