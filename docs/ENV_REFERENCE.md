# Environment Variables Reference

All configurable values for the API. Copy `.env.example` to `.env` and set as needed.

---

## Quick Reference

| Variable | Default | Description |
|----------|---------|--------------|
| `APP_ENV` | development | `production` = strict JWT secret check |
| `HTTP_ADDR` | :8080 | Listen address |
| `HTTP_READ_TIMEOUT` | 10s | Server read timeout |
| `HTTP_WRITE_TIMEOUT` | 30s | Server write timeout |
| `HTTP_IDLE_TIMEOUT` | 1m | Server idle timeout |
| `REQUEST_TIMEOUT` | 60s | Max time per request (middleware) |
| `SHUTDOWN_TIMEOUT` | 10s | Graceful shutdown wait |
| `JWT_SECRET` | (dev default) | **32+ chars in production** |
| `JWT_EXPIRY_HOURS` | 24 | Token validity in hours |
| `JWT_ISSUER` | ecom-api | iss claim |
| `JWT_AUDIENCE` | ecom-api | aud claim |
| `RATE_LIMIT_REQUESTS_PER_MINUTE` | 100 | Per-IP limit (0 = disabled) |
| `CORS_ALLOWED_ORIGINS` | (empty) | Comma-separated origins |
| `GOOSE_DBSTRING` | — | Full DB connection string |
| `DB_HOST` | localhost | (when GOOSE_DBSTRING unset) |
| `DB_PORT` | 15432 | (when GOOSE_DBSTRING unset) |
| `DB_USER` | postgres | |
| `DB_PASSWORD` | postgres | |
| `DB_NAME` | ecom | |
| `DB_POOL_MAX_CONNS` | 25 | Connection pool max |
| `DB_POOL_MIN_CONNS` | 2 | Connection pool min |

---

## Duration Format

For `HTTP_*_TIMEOUT`, `REQUEST_TIMEOUT`, `SHUTDOWN_TIMEOUT`:

- `10s` = 10 seconds
- `1m` = 1 minute
- `24h` = 24 hours

Or integer seconds: `60` = 60 seconds.

---

## Production Checklist

- `APP_ENV=production`
- `JWT_SECRET` = output of `openssl rand -base64 32`
- Strong `DB_PASSWORD` (or use `GOOSE_DBSTRING` with managed DB)
- `CORS_ALLOWED_ORIGINS` = your frontend URL(s)
