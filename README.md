# Teeth photo bot (Telegram + Gemini Vision)

Go service that receives Telegram **webhook** updates, downloads tooth photos, optionally resizes them, sends them to the **Google Gemini** API for **informational** analysis, and replies in Telegram.

**Important:** This project is intentionally **non-diagnostic**. Bot replies must be treated as **general information only**, not medical or dental advice.

## Features

- Telegram Bot API via **webhooks** (no long polling)
- Image download + resize/compress (JPEG) before Gemini
- Gemini structured JSON (`application/json`) mapped to a fixed schema
- `gin` HTTP server with middleware (request ID, structured logging, recovery, optional webhook secret)
- Context timeouts on outbound HTTP (Telegram file fetch + Gemini)
- Docker / Docker Compose for local runs
- Environment-based configuration (`godotenv` loads `.env` when present)

## Configuration

Copy `.env.example` to `.env` and fill in secrets.

| Variable | Required | Description |
|----------|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | yes | From [@BotFather](https://t.me/BotFather) |
| `GEMINI_API_KEY` | yes | Google AI Studio / Gemini API key |
| `GEMINI_MODEL` | no | Default `gemini-1.5-flash` (vision-capable) |
| `TELEGRAM_WEBHOOK_SECRET` | no | If set, Telegram must send matching `X-Telegram-Bot-Api-Secret-Token` |
| `HTTP_ADDR` | no | Default `:8080` |
| `REQUEST_TIMEOUT` | no | Per-webhook processing timeout (default `55s`) |
| `GEMINI_HTTP_TIMEOUT` | no | Gemini HTTP client timeout (default `60s`) |
| `TELEGRAM_DOWNLOAD_TIMEOUT` | no | Telegram file download timeout (default `30s`) |
| `MAX_IMAGE_DIMENSION` | no | Longest edge after resize (default `1024`) |
| `LOG_LEVEL` | no | `info` or `debug` |

## Local development

### Prerequisites

- Go 1.22+
- Docker (optional, for Compose)
- A **public HTTPS URL** for Telegram to call your webhook (use ngrok, Cloudflare Tunnel, or deploy to a host with TLS)

### Run with Go

```bash
cp .env.example .env
# edit .env
make tidy
make run
```

Expose HTTPS locally (example with ngrok):

```bash
ngrok http 8080
```

Register the webhook (replace values):

```bash
export BOT_TOKEN="YOUR_TELEGRAM_BOT_TOKEN"
export PUBLIC_BASE="https://YOUR-NGROK-SUBDOMAIN.ngrok-free.app"

curl -fsS "https://api.telegram.org/bot${BOT_TOKEN}/setWebhook" \
  -d "url=${PUBLIC_BASE}/webhook" \
  -d "secret_token=YOUR_SECRET_IF_USING_TELEGRAM_WEBHOOK_SECRET"
```

If you set `TELEGRAM_WEBHOOK_SECRET` in `.env`, pass the same value as `secret_token` here.

Health check: `GET /health`.

### Run with Docker Compose

```bash
cp .env.example .env
# edit .env
make docker-up
```

Map your tunnel to `localhost:${HOST_PORT:-8080}` and `setWebhook` as above.

## Expected Gemini JSON shape

The prompt asks Gemini to return JSON only:

```json
{
  "visible_issues": ["possible plaque buildup"],
  "confidence": "low",
  "recommendations": ["brush twice daily"],
  "disclaimer": "This is not medical advice."
}
```

Replies to users include an explicit **informational-only / not medical advice** disclaimer.

## Project layout

```
cmd/teeth-bot/main.go          # HTTP server bootstrap
internal/config                # Env configuration
internal/handler               # Webhook + formatting
internal/gemini                # Gemini REST client
internal/imageproc             # Resize/compress
internal/middleware            # Gin middleware
internal/model                 # Shared structs
internal/port                  # Interfaces for testing/mocking
internal/telegrambot           # Telegram download + send
```

## Makefile targets

- `make run` — `go run ./cmd/teeth-bot`
- `make build` — binary to `bin/teeth-bot`
- `make test` — unit tests
- `make docker-up` / `make docker-down` — Compose

## Safety & compliance note

This bot must **not** be presented as a diagnostic tool. Operators are responsible for consent, privacy (health-adjacent photos), retention policies, and compliance with Telegram/Google terms.
