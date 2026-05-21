# Teeth photo bot + Telegram Mini App (Go + Gemini)

Go service that:

- Receives Telegram **webhook** updates, analyzes tooth photos with **Google Gemini**, and replies in Russian.
- Exposes a **REST API** for Telegram Mini App **authentication** (initData HMAC) and **appointment booking** (PostgreSQL).

**Important:** Analysis is **non-diagnostic**. Bot and Mini App text must not be treated as medical advice.

## Features

### Telegram bot (webhook)

- Image download, resize, Gemini structured JSON analysis
- Russian user messages (`/start`, help, errors)

### Mini App API

- `POST /auth/telegram` — validate Telegram `initData`, upsert user, issue JWT
- `POST /appointments` — book a slot (authenticated)
- `GET /appointments/me` — list own appointments (patients)
- `GET /appointments` — list all appointments with patient info (**doctors only**)

### Architecture

Clean / screaming architecture:

```
cmd/teeth-bot/          # fx bootstrap only
infra/                  # config, HTTP router, PostgreSQL pool + migrations
internal/
  domain/               # User, Appointment, typed errors
  port/                 # Repository & auth interfaces
  usecase/              # Telegram login, create/list appointments
  service/              # Date/time validation rules
  adapters/
    driving/http/       # Gin handlers, JWT middleware, DTO converters
    driven/             # Postgres, JWT, Telegram initData validator
web/miniapp/            # TypeScript Telegram Mini App UI (Russian)
```

See [docs/API.md](docs/API.md) for request/response examples.

## Configuration

Copy `.env.example` to `.env` and fill in secrets.

| Variable | Required | Description |
|----------|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | yes | From [@BotFather](https://t.me/BotFather) |
| `GEMINI_API_KEY` | yes | Gemini API key |
| `DATABASE_URL` | yes | PostgreSQL DSN |
| `JWT_SECRET` | yes | HS256 secret, **min 32 chars** |
| `JWT_TTL` | no | Access token lifetime (default `24h`) |
| `TELEGRAM_AUTH_MAX_AGE` | no | Max initData age (default `24h`) |
| `CORS_ALLOW_ORIGINS` | no | Comma-separated origins for Mini App dev |
| `DOCTOR_TELEGRAM_IDS` | no | Comma-separated Telegram user IDs granted doctor role |
| `TELEGRAM_WEBHOOK_SECRET` | no | Webhook `X-Telegram-Bot-Api-Secret-Token` |
| `HTTP_ADDR` | no | Default `:8080` |
| `GEMINI_MODEL` | no | Default `gemini-1.5-flash` |

## Local development

### Backend + database

```bash
cp .env.example .env
# set TELEGRAM_BOT_TOKEN, GEMINI_API_KEY, JWT_SECRET (32+ chars)

docker compose up -d postgres
make tidy
make run
```

Migrations run automatically on startup (`infra/postgresql/migrations`).

Health: `GET /health`.

### Telegram webhook

Expose HTTPS (ngrok / tunnel) and register:

```bash
export BOT_TOKEN="YOUR_TELEGRAM_BOT_TOKEN"
export PUBLIC_BASE="https://YOUR-TUNNEL.example"

curl -fsS "https://api.telegram.org/bot${BOT_TOKEN}/setWebhook" \
  -d "url=${PUBLIC_BASE}/webhook" \
  -d "secret_token=YOUR_SECRET_IF_SET"
```

### Mini App (frontend)

```bash
cd web/miniapp
cp .env.example .env
# VITE_API_BASE_URL=http://localhost:8080
npm install
npm run dev
```

In [@BotFather](https://t.me/BotFather), set the bot **Menu Button** / Web App URL to your hosted Mini App (HTTPS). For local dev, use a tunnel to Vite (`5173`) and add that origin to `CORS_ALLOW_ORIGINS`.

The Mini App sends `Telegram.WebApp.initData` to `POST /auth/telegram` and uses the returned JWT for appointment endpoints.

### Full stack with Docker Compose

```bash
cp .env.example .env
make docker-up
```

## Makefile

| Target | Description |
|--------|-------------|
| `make run` | Run API + bot webhook |
| `make build` | Build `bin/teeth-bot` |
| `make test` | Unit tests |
| `make miniapp-dev` | Vite dev server for Mini App |
| `make docker-up` | Compose: Postgres + API |

## Safety

Operators are responsible for consent, privacy (health-adjacent photos), retention, and compliance with Telegram/Google terms. Appointment booking does not replace clinic scheduling systems until integrated.
