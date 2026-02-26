# Issues Faced and How to Run Successfully

This document summarizes the problems encountered while setting up and running this e-commerce API project, and how each was fixed. Use it as a reference when something goes wrong or when you want to understand why things are configured the way they are.

---

## Summary of Issues

| # | Issue | Cause | Fix |
|---|--------|--------|-----|
| 1 | `role "postgres" does not exist` | App was connecting to port **5432** (another Postgres), not Docker | Use port **15432** for Docker Postgres |
| 2 | App connecting to wrong port (5432) | DSN had no port or had `port=5432`; another Postgres was on 5432 | Force port 15432 in code and in `.env` |
| 3 | Panic on DB connection failure | Code used `panic(err)` | Replaced with log + `os.Exit(1)` |
| 4 | DB credentials “without username password” | Confusion about where credentials come from | Load `.env` and build DSN from env vars with defaults |
| 5 | `.env` not applied when running app | Not loading `.env` or wrong working directory | Use `godotenv.Load(".env")` and run from project root |
| 6 | Port 8080 “already in use” (during testing) | Another process was using 8080 on that machine | Use **8080** normally; use **8081** only if 8080 is busy on your machine |

---

## 1. `role "postgres" does not exist` (SQLSTATE 28000)

### What you saw

```
panic: failed to connect to `user=postgres database=ecom`:
  [::1]:5432 (localhost): server error: FATAL: role "postgres" does not exist (SQLSTATE 28000)
  127.0.0.1:5432 (localhost): server error: FATAL: role "postgres" does not exist (SQLSTATE 28000)
```

### Why it happened

- The app was connecting to **port 5432** (PostgreSQL default).
- On your Mac, **another** PostgreSQL was already using 5432 (e.g. Postgres.app, Homebrew Postgres, or another Docker container).
- That other instance either:
  - Was not created with a user named `postgres`, or  
  - Was a different cluster, so the `postgres` role did not exist there.

So the app was talking to the **wrong** database.

### Fix

- **Docker Postgres** for this project is mapped to host port **15432** in `docker-compose.yaml` (`15432:5432`).
- The app and all tools (e.g. Goose) must use **port 15432** so they connect to the **project’s** Postgres (which has user `postgres` and database `ecom`).

**What we did:**

- In `docker-compose.yaml`: map container 5432 to host **15432**.
- In `.env`: set `GOOSE_DBSTRING=... port=15432 ...`.
- In `cmd/main.go`: default DSN uses `port=15432`, and if `GOOSE_DBSTRING` has `port=5432` or no port, we replace/add `port=15432` so the app always talks to Docker Postgres.

**What you should do:**

- Run the app from the **project root** so `.env` is loaded.
- Do **not** set `GOOSE_DBSTRING` in your shell to a DSN with `port=5432` (or with no port, on a machine where 5432 is another Postgres).

---

## 2. App still connecting to port 5432

### What you saw

Error message still showed `:5432` in the connection error (e.g. `127.0.0.1:5432`).

### Why it happened

- `GOOSE_DBSTRING` was set in the **shell** to something like `host=localhost user=postgres ...` (no port or with `port=5432`).
- Or `.env` was not loaded because the app was run from a different directory.

### Fix

- **Run from project root:**  
  `cd /Users/iniyan/Documents/Backend/ecom-go-api-project` then `go run cmd/*.go`.
- **Load `.env`:** The app calls `godotenv.Load(".env")` so variables from `.env` (with `port=15432`) are used.
- **Code-level safeguard:** In `getDSN()` we now:
  - If `GOOSE_DBSTRING` has no `port=`, we add `port=15432`.
  - If it has `port=5432`, we replace it with `port=15432`.

So even with an old or shell-set DSN, the app will use 15432 and connect to Docker.

---

## 3. Panic on database connection failure

### What you saw

```
/Users/iniyan/Documents/Backend/ecom-go-api-project/cmd/main.go:29 +0x354
exit status 2
```

Line 29 was `panic(err)` when `pgx.Connect` failed.

### Why it happened

Any DB connection error (wrong port, wrong password, Postgres down) caused a **panic** instead of a clean error message.

### Fix

- Replaced `panic(err)` with:
  - Log the error (and host/port/db) with `logger.Error(...)`.
  - Exit with `os.Exit(1)`.
- So now you get a clear log line and exit code 1, without a stack trace.

---

## 4. “Without username password how will DB connect?”

### What it meant

- You wanted to understand how the app gets DB credentials and to make it work reliably.
- The DB **does** need a username and password; they should come from config, not be hardcoded in code.

### Fix

- **Username and password** come from:
  - **Option 1:** `.env` → `GOOSE_DBSTRING=host=localhost port=15432 user=postgres password=postgres dbname=ecom sslmode=disable`.
  - **Option 2:** Separate env vars: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` (with defaults in code).
- The app loads `.env` with `godotenv` and builds the DSN in `getDSN()` so the DB **always** connects with user and password from env or defaults.

---

## 5. `.env` not applied

### What you saw

- Even with correct values in `.env`, the app sometimes used the wrong port or credentials.

### Why it happened

- Go does **not** load `.env` by default.
- If the app was run from a different directory, `godotenv.Load(".env")` might not find the file.

### Fix

- Call `godotenv.Load(".env")` at the start of `main()`.
- **Always run the app from the project root:**  
  `cd /Users/iniyan/Documents/Backend/ecom-go-api-project` then `go run cmd/*.go` (or `./scripts/setup-and-run.sh`).

---

## 6. Port 8080 vs 8081

### What you saw

- In some instructions or tests you saw **8081** as well as **8080**.

### Why 8081 was mentioned

- During automated testing, **another process** on that machine was already using port **8080** (e.g. a previous run of the same API or another app).
- So the test used `HTTP_ADDR=:8081` to avoid “address already in use” and to prove the API and DB work.

### How it works for you

- **8080 is the default.** The app is meant to run on **http://localhost:8080**.
- Use **8081 only if** on your machine you get:
  ```text
  listen tcp :8080: bind: address already in use
  ```
- Then run:
  ```bash
  HTTP_ADDR=:8081 go run cmd/*.go
  ```
  and use **http://localhost:8081** in Postman instead of 8080.

So: **use 8080 normally; 8081 is only a fallback when 8080 is already in use.**

---

## How to Run Successfully (Checklist)

1. **Install once:** Go, Docker (and Docker Compose), Goose, SQLC (see main [README.md](README.md)).
2. **Start Postgres:**  
   `docker compose up -d`  
   (Postgres will be on **port 15432** on the host.)
3. **Run from project root:**  
   `cd /Users/iniyan/Documents/Backend/ecom-go-api-project`
4. **Migrations (first time or after schema change):**  
   ```bash
   export GOOSE_DRIVER=postgres
   export GOOSE_DBSTRING="host=localhost port=15432 user=postgres password=postgres dbname=ecom sslmode=disable"
   export GOOSE_MIGRATION_DIR=internal/adapters/postgresql/migrations
   goose up
   ```
5. **Start the API:**  
   `go run cmd/*.go`  
   You should see something like:  
   `level=INFO msg="connected to database" host=localhost port=15432 db=ecom`  
   `server has started at addr :8080`
6. **Test in Postman:**  
   - **GET** `http://localhost:8080/health` → `all good`  
   - **GET** `http://localhost:8080/products` → JSON list of products  
   - **POST** `http://localhost:8080/orders` with `Content-Type: application/json` and body:  
     `{"customerId": 1, "items": [{"productId": 1, "quantity": 2}]}`

**One-command option:** From project root run `./scripts/setup-and-run.sh`; it starts Postgres, runs migrations, seeds products, and starts the app on **8080** (use 8081 only if 8080 is in use on your machine).

---

## Docker has its own Postgres (not your Mac’s)

**Short answer:** Docker runs **its own** PostgreSQL inside the container. It does **not** use the PostgreSQL installed on your Mac (e.g. Postgres.app or Homebrew).

- **Docker Postgres:** Lives inside the `ecom-postgres` container. Data is stored in a Docker volume (`postgres-data`). You connect to it from your Mac at **localhost:15432** (because `docker-compose.yaml` maps container port 5432 to host port 15432).
- **Your Mac Postgres:** If you have Postgres.app or `brew install postgresql`, that runs **separately** on your machine, often on port **5432**. It’s a different server, different data. The API and this project are set up to use the **Docker** one (15432), not the Mac one (5432).

So: two separate PostgreSQLs. The project uses only the one in Docker (15432).

---

## Connection test failing when I use port 15432

If your DB client (TablePlus, DBeaver, pgAdmin, etc.) or an app “connection test” fails when you use port **15432**, check the following.

1. **Is the Docker container running?**
   ```bash
   docker ps
   ```
   You should see `ecom-postgres` with status “Up”. If not:
   ```bash
   cd /path/to/ecom-go-api-project
   docker compose up -d
   ```

2. **Exact connection settings (must match):**
   - **Host:** `localhost` (or `127.0.0.1`)
   - **Port:** `15432` (number, not 5432)
   - **Database:** `ecom`
   - **User:** `postgres`
   - **Password:** `postgres`
   - **SSL/TLS:** Turn **off** or use “prefer” / “disable” (the project uses `sslmode=disable` in the DSN). Some clients fail the test if they insist on SSL and the server doesn’t require it.

3. **Test from the terminal (bypasses the client):**
   ```bash
   docker exec ecom-postgres psql -U postgres -d ecom -c "SELECT 1;"
   ```
   If this works, Postgres inside Docker is fine; the issue is likely host/port/SSL in your client.

4. **Firewall:** Rare, but if you’re on a locked-down network, ensure nothing is blocking local port 15432.

Once the connection test passes with **port 15432** and **SSL off**, you’re talking to the same database the API uses.

---

## Quick Reference: Ports

| Port  | Use |
|-------|------|
| **15432** | PostgreSQL (Docker) on the host. App and Goose must use this so they hit the project’s DB. |
| **5432**  | Default Postgres port; often used by another instance on the Mac. Do **not** use this for this project. |
| **8080**  | Default HTTP port for this API. Use this in Postman. |
| **8081**  | Use only if 8080 is already in use; then set `HTTP_ADDR=:8081` and use `http://localhost:8081` in Postman. |

This file is for your own reference to remember what went wrong and how to make the project run successfully.
