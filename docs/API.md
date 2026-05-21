# REST API

Base URL: `http://localhost:8080` (or your deployed host).

All error responses:

```json
{ "code": "invalid_init_data", "message": "invalid telegram init data" }
```

## POST /auth/telegram

Authenticates a Telegram Mini App user using signed `initData`.

**Request**

```json
{
  "init_data": "query_id=...&user=%7B%22id%22%3A...%7D&auth_date=...&hash=..."
}
```

**Response `200`**

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2026-05-21T12:00:00Z",
  "user": {
    "id": 1,
    "telegram_id": 123456789,
    "username": "ivan",
    "first_name": "Иван",
    "last_name": "Иванов",
    "avatar_url": "https://..."
  }
}
```

**Example**

```bash
curl -sS -X POST http://localhost:8080/auth/telegram \
  -H 'Content-Type: application/json' \
  -d '{"init_data":"'"$TELEGRAM_INIT_DATA"'"}'
```

## POST /appointments

Creates an appointment for the authenticated user.

**Headers**

`Authorization: Bearer <access_token>`

**Request**

```json
{
  "preferred_date": "2026-05-25",
  "preferred_time": "14:30"
}
```

- `preferred_date`: `YYYY-MM-DD`, today … +90 days (UTC)
- `preferred_time`: `HH:MM`, clinic hours **09:00–20:00**

**Response `201`**

```json
{
  "id": 1,
  "preferred_date": "2026-05-25",
  "preferred_time": "14:30",
  "status": "pending",
  "created_at": "2026-05-20T10:00:00Z"
}
```

**Example**

```bash
curl -sS -X POST http://localhost:8080/appointments \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"preferred_date":"2026-05-25","preferred_time":"14:30"}'
```

## GET /appointments

Lists **all** patient appointments with patient details. **Doctors only** (see `DOCTOR_TELEGRAM_IDS`).

**Headers**

`Authorization: Bearer <doctor_access_token>`

**Response `200`**

```json
{
  "appointments": [
    {
      "id": 1,
      "preferred_date": "2026-05-25",
      "preferred_time": "14:30",
      "status": "pending",
      "created_at": "2026-05-20T10:00:00Z",
      "patient": {
        "id": 2,
        "telegram_id": 987654321,
        "username": "maria",
        "first_name": "Мария",
        "last_name": "Петрова"
      }
    }
  ]
}
```

**Example**

```bash
curl -sS http://localhost:8080/appointments \
  -H "Authorization: Bearer $DOCTOR_TOKEN"
```

Configure doctors in `.env`:

```env
DOCTOR_TELEGRAM_IDS=123456789,987654321
```

After the next Mini App login, those Telegram accounts receive `role: doctor` and the doctor dashboard in the Mini App.

## GET /appointments/me

Lists appointments for the authenticated user.

**Headers**

`Authorization: Bearer <access_token>`

**Response `200`**

```json
{
  "appointments": [
    {
      "id": 1,
      "preferred_date": "2026-05-25",
      "preferred_time": "14:30",
      "status": "pending",
      "created_at": "2026-05-20T10:00:00Z"
    }
  ]
}
```

**Example**

```bash
curl -sS http://localhost:8080/appointments/me \
  -H "Authorization: Bearer $TOKEN"
```

## Database schema

See `migrations/001_init.up.sql` and `infra/postgresql/migrations/001_init.up.sql`.

| Table | Columns |
|-------|---------|
| `users` | `id`, `telegram_id`, `username`, `first_name`, `last_name`, `avatar_url`, `created_at`, `updated_at` |
| `appointments` | `id`, `user_id`, `preferred_date`, `preferred_time`, `status`, `created_at` |
