# How to Run the Ecom API — Final Steps

Follow these steps in order. No prior setup needed if you have Docker and Go installed.

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

# Get products
curl "http://localhost:8080/products?limit=5"

# Place order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customerId":1,"items":[{"productId":1,"quantity":2}]}'
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

## If Something Fails

| Error | Fix |
|-------|-----|
| "Cannot connect to Docker daemon" | Start Docker Desktop |
| "port 8080 already in use" | Stop other app on 8080, or use `HTTP_ADDR=:8081 go run ./cmd` |
| "role postgres does not exist" | See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) |
| **Register/login not working** | Use `POST /v1/auth/register` and `POST /v1/auth/login`. Add `JWT_SECRET` to `.env` (32+ chars). |
