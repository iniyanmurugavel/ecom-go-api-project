# Issues Faced and How to Run Successfully

This document summarizes the problems encountered while setting up and running this e-commerce API project, and how each was fixed. Use it as a reference when something goes wrong or when you want to understand why things are configured the way they are.

**New to Go?** See [docs/LEARNER_GUIDE.md](docs/LEARNER_GUIDE.md) for concepts and study path.

---

## Quick Reference: Ports and Connections

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Your machine                                                                │
│                                                                              │
│  API (go run ./cmd)          Docker Postgres                                 │
│  listens on :8080            listens on host port 15432                      │
│       │                              ▲                                        │
│       │                              │                                        │
│       │  connects via DSN            │  docker compose maps                   │
│       │  host=localhost port=15432   │  15432 (host) → 5432 (container)       │
│       └─────────────────────────────┘                                        │
│                                                                              │
│  Mac Postgres (if installed) often uses 5432 — we use 15432 to avoid clash.  │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Summary of Issues

| # | Issue | Cause | Fix |
|---|--------|--------|-----|
| 1 | `role "postgres" does not exist` | Connected to Mac Postgres (5432) instead of Docker (15432) | Use **port 15432** in `.env` and DB client for Docker Postgres |
| 2 | App connecting to wrong port | DSN had wrong port or `.env` not loaded | Check `.env` has `port=15432`; run from project root |
| 3 | Panic on DB connection failure | Code used `panic(err)` | Replaced with log + `os.Exit(1)` |
| 4 | DB credentials “without username password” | Confusion about where credentials come from | Load `.env` and build DSN from env vars with defaults |
| 5 | `.env` not applied when running app | Not loading `.env` or wrong working directory | Use `godotenv.Load(".env")` and run from project root |
| 6 | Port 8080 "already in use" (during testing) | Another process was using 8080 on that machine | Use **8080** normally; use **8081** only if 8080 is busy on your machine |
| 7 | Xcode license / VCS error | macOS: CGO or git needs Xcode | Use `CGO_ENABLED=0 go run -buildvcs=false ./cmd` — see [Section 7](#7-xcode-license-or-vcs-error-macos) |


---

## 1. `role "postgres" does not exist` (SQLSTATE 28000)

### What you saw

```
panic: failed to connect to `user=postgres database=ecom`:
  [::1]:5432 (localhost): server error: FATAL: role "postgres" does not exist (SQLSTATE 28000)
  127.0.0.1:5432 (localhost): server error: FATAL: role "postgres" does not exist (SQLSTATE 28000)
```

### Why it happened

- The app was connecting to **port 5432** (Mac Postgres), not Docker on 15432.
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

**What you should do:**

- Run the app from the **project root** so `.env` is loaded.
- Port 15432 avoids conflict with Mac Postgres on 5432.

---

## 2. App still connecting to wrong port

### What you saw

Error message still showed `:5432` in the connection error (e.g. `127.0.0.1:5432`).

### Why it happened

- `GOOSE_DBSTRING` was set in the **shell** to something like `host=localhost user=postgres ...` (no port or with `port=5432`).
- Or `.env` was not loaded because the app was run from a different directory.

### Fix

- **Run from project root:**  
  `cd /Users/iniyan/Documents/Backend/ecom-go-api-project` then `go run ./cmd`.
- **Load `.env`:** The app calls `godotenv.Load(".env")` so variables from `.env` (with `port=15432`) are used.

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
  `cd /Users/iniyan/Documents/Backend/ecom-go-api-project` then `go run ./cmd` (or `./scripts/setup-and-run.sh`).

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
  HTTP_ADDR=:8081 go run ./cmd
  ```
  and use **http://localhost:8081** in Postman instead of 8080.

So: **use 8080 normally; 8081 is only a fallback when 8080 is already in use.**

---

## 7. Xcode License or VCS Error (macOS)

### What you saw

```
# runtime/cgo
You have not agreed to the Xcode license agreements. Please run 'sudo xcodebuild -license'
```

or

```
error obtaining VCS status: exit status 69
Use -buildvcs=false to disable VCS stamping.
```

### Why it happened

- **CGO:** Go can call C code. On macOS, that uses Xcode tooling (clang). If the Xcode license isn't accepted, the build fails.
- **VCS:** Go embeds git info (commit, dirty flag) in binaries. If git fails (e.g. due to Xcode license when running hooks), the build fails.

### Fix

Run with:

```bash
CGO_ENABLED=0 go run -buildvcs=false ./cmd
```

| Part | What it does |
|------|--------------|
| `CGO_ENABLED=0` | Disables CGO. No C code, no Xcode. This project uses pure-Go pgx, so CGO isn't needed. |
| `-buildvcs=false` | Skips embedding git info. Avoids VCS-related build errors. |
| `./cmd` | The package to build and run. |

**Alternative:** Accept the Xcode license: `sudo xcodebuild -license`, then `go run ./cmd` works without these flags.

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
   `go run ./cmd`  
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

So: this project has **one database** — the one in Docker (port 15432). Your Mac may have another Postgres on 5432; we do not use that.

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

2. **Exact connection settings (must match)** — same as README "Standard Docker Postgres connection":

   | Host     | Port   | Database | User     | Password | SSL  |
   |----------|--------|----------|----------|----------|------|
   | localhost | 15432 | ecom     | postgres | postgres | off  |

   Turn **SSL/TLS** off; the project uses `sslmode=disable`. Some clients fail if they insist on SSL.

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
| **8080**  | Default HTTP port for this API. Use this in Postman. |
| **8081**  | Use only if 8080 is already in use; then set `HTTP_ADDR=:8081` and use `http://localhost:8081` in Postman. |

This file is for your own reference to remember what went wrong and how to make the project run successfully.
