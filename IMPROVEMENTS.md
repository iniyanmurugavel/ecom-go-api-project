# Improvements Made: Scalable REST API Practices

This document summarizes all changes made to align the project with scalable REST API practices. Use it to understand what was improved and why, so you can apply similar patterns in future projects.

---

## 1. What Was Improved (Summary)

| Area | Before | After |
|------|--------|-------|
| **Database connection** | Single `*pgx.Conn` (one request at a time) | `pgxpool.Pool` (many concurrent requests) |
| **Repository layer** | Services depended on concrete `*repo.Queries` | Services depend on interfaces (`Repository`, `OrderRepo`, `TxBeginner`) |
| **Logger** | Not passed; used global `log` | Injected `*slog.Logger` into handlers and services |
| **API routes** | `/products`, `/orders` | `/v1/products`, `/v1/orders` (versioned) |
| **List products** | No pagination | `?limit=20&offset=0` (default 20, max 100) |
| **Health check** | `/health` only | `/health` + `/health/live` (pings DB; 503 if down) |
| **Shutdown** | Ctrl+C killed process immediately | Graceful shutdown: waits for in-flight requests (up to 10s) |
| **Domain errors** | Some mapped to HTTP status | All mapped: 400, 404, 409, 500 |
| **Rate limiting** | None | Per-IP limit (100/min default; configurable via `RATE_LIMIT_REQUESTS_PER_MINUTE`) |
| **Unit tests** | None | Products and orders services tested with mock repos |
| **Integration tests** | None | HTTP endpoints tested against real DB (skips if DB unavailable) |
| **CI** | None | GitHub Actions: migrations, build, unit + integration tests |
| **Documentation** | Basic | One DB clarified; stop commands; consistent connection table; comments in code |

---

## 2. Files Changed or Added

### New files

- `internal/products/repository.go` – `Repository` interface for products
- `internal/orders/repository.go` – `OrderRepo`, `OrderTxRepo`, `TxBeginner` interfaces; `NewOrderRepo` adapter
- `internal/products/service_test.go` – Unit tests with mock repo
- `internal/orders/service_test.go` – Unit tests with mock repo, fake `pgx.Tx`
- `cmd/integration_test.go` – Integration tests (HTTP + real DB)
- `.github/workflows/ci.yml` – CI: Postgres service, goose up, build, tests
- `IMPROVEMENTS.md` – This file

### Modified files

- `cmd/main.go` – pgxpool, graceful shutdown, rate limit config, logger
- `cmd/api.go` – Pool, logger, `/v1` routes, health/live, rate limit middleware, comments
- `internal/products/service.go` – Repository interface, logger, pagination params
- `internal/products/handlers.go` – Logger, pagination (limit/offset), comments
- `internal/orders/service.go` – OrderRepo + TxBeginner, logger, `ErrInvalidInput`
- `internal/orders/handlers.go` – Logger, map `ErrInvalidInput` → 400, `ErrProductNoStock` → 409
- `internal/env/env.go` – `GetInt` for rate limit
- `internal/adapters/postgresql/sqlc/queries.sql` – `ListProductsPaginated` query
- `go.mod` / `go.sum` – pgxpool, httprate, test deps
- `docker-compose.yaml` – Comments (one DB, start/stop, port mapping)
- `.env.example` – `RATE_LIMIT_REQUESTS_PER_MINUTE`
- `README.md` – /v1, health/live, pagination, stop commands, scalability section, Postman URLs
- `ARCHITECTURE.md` – /v1 routes, pagination flow
- `scripts/setup-and-run.sh` – /v1 in Postman hint

---

## 3. README Updates

The main README was updated to reflect:

- **Endpoints**: `GET /health`, `GET /health/live`, `GET /v1/products`, `POST /v1/orders`
- **Pagination**: `?limit=20&offset=0` for products
- **One database**: Clarified that API, Goose, and DB clients all use the same Docker Postgres
- **Stop commands**: Table for stopping API (Ctrl+C) and DB (`docker compose down` / `down -v`)
- **Standard connection table**: Host, Port, Database, User, Password, SSL (consistent across docs)
- **Scalability section**: Connection pool, repository interfaces, versioning, pagination, graceful shutdown, health/live, rate limiting
- **Postman URLs**: All use `/v1/products` and `/v1/orders`

---

## 4. Why These Changes Matter (For Learning)

1. **Connection pool** – Handles many requests at once instead of one at a time.
2. **Repository interfaces** – Lets you test services without a real DB (mock the repo).
3. **Logger injection** – Centralized, structured logging for debugging and production.
4. **API versioning** – Lets you add `/v2` later without breaking existing clients.
5. **Pagination** – Keeps responses bounded and prepares for growth.
6. **Graceful shutdown** – Avoids dropping requests during deploys.
7. **Health with DB** – Load balancers can stop sending traffic to unhealthy instances.
8. **Domain errors → HTTP status** – Clients get stable, predictable status codes.
9. **Rate limiting** – Protects the API and DB from abuse.
10. **Unit + integration tests** – Catch regressions and validate migrations in CI.

---

## 5. How to Run After These Changes

Same as before, but use the new URLs in Postman:

```bash
# 1. Start DB
docker compose up -d

# 2. Migrate (if first time or new migrations)
source .env && goose up

# 3. Run API
go run cmd/*.go
```

**Postman:**

- `GET http://localhost:8080/health`
- `GET http://localhost:8080/health/live`
- `GET http://localhost:8080/v1/products?limit=20&offset=0`
- `POST http://localhost:8080/v1/orders` with JSON body

---

## 6. Branch and Push

These improvements are on the `feature/scalable-improvements` branch. To push:

```bash
git checkout feature/scalable-improvements
git push -u origin feature/scalable-improvements
```

Then open a Pull Request on GitHub to merge into `main` when ready.
