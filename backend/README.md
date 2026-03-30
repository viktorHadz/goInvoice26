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

For local development the callback URL should usually be:

```text
http://localhost:4206/api/auth/google/callback
```

That same callback URI must be added to your Google OAuth app configuration.
