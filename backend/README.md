# GoInvoicer (Go + SQLite)

![goInvoicer Mascot](./goInvoicerMascot.png)

## Requirements

- Go 1.24+ (or your chosen version)

## Local development

```bash
cp .env.example .env
mkdir -p data
make run
```

## Google auth setup

Owner signup and login now use Google OAuth and a server-side session cookie.

Set these values in `.env`:

- `APP_BASE_URL`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GOOGLE_REDIRECT_URL`
- `STRIPE_SECRET_KEY`
- `STRIPE_PRICE_ID`
- `STRIPE_WEBHOOK_SECRET`

For local development the callback URL should usually be:

```text
http://localhost:4206/api/auth/google/callback
```

That same callback URI must be added to your Google OAuth app configuration.

## Stripe billing setup

Billing is account-scoped and uses Stripe Checkout plus a webhook to keep subscription
state in sync.

Set these values in `.env`:

- `STRIPE_PUBLISHABLE_KEY`
- `STRIPE_SECRET_KEY`
- `STRIPE_PRICE_ID`
- `STRIPE_WEBHOOK_SECRET`

For local development, point your Stripe webhook endpoint at:

```text
http://localhost:4206/api/billing/stripe/webhook
```

The hosted checkout flow returns users to:

```text
http://localhost:5173/app/billing
```
