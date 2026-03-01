# How to Run and Test the API

Single reference for running the API and running tests (unit + integration).

---

## 1. How to Run the API

### Quick start (recommended)

```bash
./scripts/setup-and-run.sh
```

This will:
- Start PostgreSQL in Docker
- Run migrations
- Seed sample products
- Start the API at `http://localhost:8080`

### Manual steps

```bash
# 1. Start Postgres
docker compose up -d

# 2. Run migrations (requires goose; set GOOSE_* in .env)
source .env && goose up

# 3. Start the API
go run ./cmd
```

### Verify it works

```bash
./scripts/verify-api.sh
```

Or manually: `curl http://localhost:8080/health` → should return `all good`.

---

## 2. How to Run Tests

### Unit tests (no DB required for most)

```bash
go test ./internal/... -v
```

Runs tests in `internal/auth`, `internal/orders`, `internal/products`, etc. These use mocks, so they don't need a real database.

### Integration tests (require DB)

Integration tests hit real HTTP handlers with a real Postgres database. They live in `cmd/integration_test.go`.

**Prerequisites:**
- Postgres running (e.g. `docker compose up -d`)
- `.env` with `GOOSE_DBSTRING` (or equivalent DB connection)

**Run integration tests:**

```bash
# Run all tests (unit + integration). Integration tests skip if DB is unavailable.
go test -v ./...

# Run only integration tests
go test -v ./cmd/ -run Integration
```

**What they test:**
- `TestIntegration_Health` — GET /health
- `TestIntegration_HealthLive` — GET /health/live (DB ping)
- `TestIntegration_RegisterLoginProductsOrder` — Full flow: register → login → products → order

**If DB is down:** Integration tests are skipped with a message like `integration test skipped: no DB`.

---

## 3. Quick Reference

| Task | Command |
|------|---------|
| Run API | `./scripts/setup-and-run.sh` |
| Verify API | `./scripts/verify-api.sh` |
| Unit tests | `go test ./internal/... -v` |
| Integration tests | `go test -v ./cmd/ -run Integration` |
| All tests | `go test -v ./...` |

---

## 4. Related Docs

- [HOW_TO_RUN.md](HOW_TO_RUN.md) — Step-by-step run guide
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) — Common issues
- [docs/API_REFERENCE.md](docs/API_REFERENCE.md) — All endpoints with curl
