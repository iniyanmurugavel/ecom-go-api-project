# How to Run the Ecom API — Final Steps

Follow these steps in order. No prior setup needed if you have Docker and Go installed.

**Run + tests in one place:** See [RUN_AND_TEST.md](RUN_AND_TEST.md) for run commands and integration test instructions.

**New to Go?** See [docs/LEARNER_GUIDE.md](docs/LEARNER_GUIDE.md) for a learning path and diagrams.

---

## Prerequisites

- **Docker Desktop** — must be **running** (check the menu bar icon)
- **Go** — `go version` works in terminal

---

## Step 1: Start Docker

1. Open **Docker Desktop** from Applications
2. Wait until it shows "Docker Desktop is running" (whale icon in menu bar)
3. If Docker isn't installed: [docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop/)

---

## Step 2: Open Terminal

Open **Terminal** (or iTerm2) and go to the project folder:

```bash
cd /Users/iniyan/Documents/Backend/ecom-go-api-project
```

---

## Step 3: Create .env (first time only)

```bash
cp .env.example .env
```

(Edit `.env` only if you need a different DB port or password.)

---

## Step 4: Run the Project

```bash
./scripts/setup-and-run.sh
```

This will:
- Start PostgreSQL in Docker
- Run database migrations
- Seed sample products
- Start the API on `http://localhost:8080`

You should see:
```
==> Loaded .env
==> Starting PostgreSQL...
==> Running migrations...
==> Seeding sample products...
==> Starting API server at http://localhost:8080
time=... level=INFO msg="server started" addr=:8080
```

**Keep this terminal open.** The API runs here.

---

## Step 5: Test the API

**Option A: curl (in a new terminal)**

```bash
# Health check
curl http://localhost:8080/health
# → all good

# Get products (requires token — register/login first)
curl "http://localhost:8080/v1/products?limit=5" -H "Authorization: Bearer YOUR_TOKEN"

# Place order (customer from JWT, not body)
curl -X POST http://localhost:8080/v1/orders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"productId":1,"quantity":2}]}'
```

**Option B: Postman**

1. Import `postman/Ecom-API.postman_collection.json` and `postman/Local.postman_environment.json`
2. Select **Local** environment
3. Run: **Health Check** → **Get Products** → **Place Order**

---

## Step 6: Stop the Project

- **Stop API:** Press `Ctrl+C` in the terminal where it's running
- **Stop DB:** `docker compose down` (from project folder)

---

## Quick Reference

| Action        | Command                          |
|---------------|-----------------------------------|
| Run project   | `./scripts/setup-and-run.sh`      |
| Verify API    | `./scripts/verify-api.sh` (API must be running) |
| Stop API      | `Ctrl+C`                          |
| Stop DB       | `docker compose down`             |
| Health check  | `curl http://localhost:8080/health`|

---

## Verify All Endpoints Work

Run these in a **new terminal** (with API running):

```bash
# 1. Health (should return "all good")
curl http://localhost:8080/health

# 2. Register (creates customer)
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","name":"Test User"}'

# 3. Login (get token - copy it from response)
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# 4. Products (requires token)
curl "http://localhost:8080/v1/products?limit=5" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 5. Orders (requires token; customer from JWT)
curl -X POST http://localhost:8080/v1/orders \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"productId":1,"quantity":1}]}'
```

**Auth:** Use `POST /v1/auth/register` and `POST /v1/auth/login`. Products and orders require `Authorization: Bearer <token>`.

---

## Connect a DB Client (TablePlus, DBeaver, etc.)

| Host     | Port   | Database | User     | Password | SSL  |
|----------|--------|----------|----------|----------|------|
| localhost | 15432 | ecom     | postgres | postgres | off  |

**Important:** Use **port 15432**, not 5432. Port 5432 is often used by Mac Postgres (which doesn't have the `postgres` role).

---

## If Something Fails

| Error | Fix |
|-------|-----|
| "Cannot connect to Docker daemon" | Start Docker Desktop |
| "port 8080 already in use" | Stop other app on 8080, or use `HTTP_ADDR=:8081 go run ./cmd` |
| "role postgres does not exist" | Use **port 15432** in your DB client (not 5432). See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) |
| **Register/login not working** | Use `POST /v1/auth/register` and `POST /v1/auth/login`. Add `JWT_SECRET` to `.env` (32+ chars). |
