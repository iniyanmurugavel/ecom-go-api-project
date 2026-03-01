# Deployment Guide — Production & Local Options

This document explains **where to deploy**, **how to run in production** (with and without Docker), and **local options** including running everything in Docker. No code — concepts and steps only.

---

## 1. What You Might Have Missed (Quick Checklist)

| Area | Status | Notes |
|------|--------|-------|
| Stock decrement | ✅ Done | Fixed in orders service |
| FK on order_items | ✅ Done | Migration 00004 |
| Input validation | ✅ Done | validate package |
| CORS | ✅ Done | CORS_ALLOWED_ORIGINS |
| JSON error format | ✅ Done | json.WriteError |
| Connection pool config | ✅ Done | DB_POOL_MAX_CONNS, MIN_CONNS |
| Prometheus metrics | ✅ Done | /metrics endpoint |
| OpenAPI spec | ✅ Done | docs/openapi.yaml |
| **Production env vars** | ⚠️ Set these | JWT_SECRET (32+ chars), strong DB password |
| **HTTPS / reverse proxy** | ❌ Not in app | Use nginx, Caddy, or cloud LB |
| **Secrets management** | ❌ Not in app | Use env vars, vault, or cloud secrets |

---

## 2. Where to Deploy (Options)

| Platform | Best for | Notes |
|----------|----------|-------|
| **VPS** (DigitalOcean, Linode, Vultr, AWS EC2) | Full control, low cost | You manage OS, firewall, DB, app. |
| **PaaS** (Railway, Render, Fly.io, Heroku) | Easiest deploy | Connect repo, set env vars, auto-deploy. |
| **Kubernetes** (GKE, EKS, AKS) | Large scale | Overkill for this project. |
| **Serverless** (AWS Lambda + RDS) | Event-driven | Requires adapting the app (cold starts, connection pooling). |

**Recommendation for this project:** Start with **Railway**, **Render**, or a **VPS** (DigitalOcean droplet). All support Go + Postgres.

---

## 3. Local: Can You Run Everything in Docker?

**Yes.** You have two local setups:

### Option A: DB in Docker, API on host (current default)

- **Postgres** → Docker (`docker compose up -d`)
- **API** → Your machine (`go run ./cmd`)

**Pros:** Fast edit-run cycle, easy debugging.  
**Cons:** Need Go installed; API not containerized.

### Option B: Both DB and API in Docker

- **Postgres** → Docker
- **API** → Docker (build Go binary, run in container)

**Pros:** Same as production; no Go on host.  
**Cons:** Slower feedback (rebuild image on code change); need Dockerfile.

**How it works:** Add an `api` service to `docker-compose.yaml` that builds the Go binary and runs it. The API connects to Postgres via the Docker network (hostname `postgres`, port `5432`). You’d run `docker compose up` and both start. The API would not use `localhost:15432` — it would use `postgres:5432` because they’re on the same Docker network.

**Summary:** Local with “full Docker” is possible. The project is set up for Option A; Option B needs a Dockerfile and a second compose service.

---

## 4. Production Without Docker

**Idea:** Run the Go binary on a server. Use a managed Postgres (e.g. Supabase, Neon, AWS RDS) or Postgres installed on the same/different server.

### Steps (conceptual)

1. **Build the binary:** `go build -o api ./cmd` (on your machine or in CI).
2. **Provision a server:** VPS or PaaS.
3. **Provision Postgres:** Managed DB or install Postgres on a server.
4. **Set env vars:** `GOOSE_DBSTRING`, `JWT_SECRET`, `HTTP_ADDR`, etc.
5. **Run migrations:** `goose up` against the production DB.
6. **Run the binary:** `./api` (or use systemd/supervisor to keep it running).
7. **Put a reverse proxy in front:** nginx or Caddy for HTTPS, optional rate limiting.

### Env vars for production (without Docker)

```
GOOSE_DBSTRING=host=your-db-host port=5432 user=produser password=STRONG_PASSWORD dbname=ecom sslmode=require
JWT_SECRET=at-least-32-characters-long-random-string
HTTP_ADDR=:8080
RATE_LIMIT_REQUESTS_PER_MINUTE=100
CORS_ALLOWED_ORIGINS=https://your-frontend.com
DB_POOL_MAX_CONNS=25
DB_POOL_MIN_CONNS=2
```

### Pros and cons

| Pros | Cons |
|------|------|
| Simple, no containers | You manage OS, updates, process manager |
| Works on any Linux server | Need to install Go for builds (or cross-compile) |
| Managed DB = less DB ops | |

---

## 5. Production With Docker

**Idea:** Run both API and Postgres in containers. Use `docker compose` or an orchestrator (e.g. Docker Swarm, Kubernetes).

### Option A: Docker Compose on a single server

1. **Dockerfile** for the API (build Go binary, run it).
2. **docker-compose** with `postgres` and `api` services.
3. API connects to `postgres:5432` (Docker network).
4. Expose API port (e.g. 8080) to the host.
5. Reverse proxy (nginx/Caddy) on host for HTTPS.

### Option B: PaaS with Docker (Railway, Render, Fly.io)

1. Add a Dockerfile to the repo.
2. Connect the repo to the platform.
3. Add Postgres (managed by the platform or your own).
4. Set env vars (DB URL, JWT_SECRET, etc.).
5. Deploy; platform builds and runs the container.

### Env vars for production (with Docker)

Same as above, but `GOOSE_DBSTRING` (and app DSN) use the **Postgres service name** and **internal port** when both run in Docker:

```
GOOSE_DBSTRING=host=postgres port=5432 user=postgres password=STRONG_PASSWORD dbname=ecom sslmode=disable
```

(Inside Docker network, `postgres` is the hostname. `sslmode=disable` is common inside a private network; use `require` if you add TLS.)

### Pros and cons

| Pros | Cons |
|------|------|
| Same setup locally and in prod | Need to learn Docker/Dockerfile |
| Easy to scale (more containers) | Slightly more moving parts |
| PaaS can auto-deploy from git | |

---

## 6. Production Checklist (Regardless of Setup)

| Item | Why |
|------|-----|
| **Strong JWT_SECRET** | 32+ random chars; never default |
| **Strong DB password** | Never `postgres`/`postgres` |
| **HTTPS** | Use nginx, Caddy, or cloud LB |
| **Run migrations** | `goose up` before first deploy and after schema changes |
| **Health checks** | Use `/health/live` for load balancer |
| **Logs** | Ensure stdout/stderr are captured (e.g. systemd, Docker, PaaS) |
| **Backups** | For Postgres (managed DB often includes this) |
| **CORS** | Set `CORS_ALLOWED_ORIGINS` to your frontend URL |

---

## 7. Quick Reference: Local vs Production

| | Local (default) | Local (full Docker) | Production (no Docker) | Production (Docker) |
|-|-----------------|---------------------|-------------------------|---------------------|
| **DB** | Docker | Docker | Managed or host | Docker or managed |
| **API** | `go run ./cmd` | Docker | Binary on server | Docker |
| **DB host** | localhost:15432 | postgres:5432 | your-db-host:5432 | postgres:5432 (or managed) |
| **Build** | go run | docker build | go build | docker build |

---

## 8. Related Docs

- [DEVELOPMENT_AND_DEPLOYMENT.md](DEVELOPMENT_AND_DEPLOYMENT.md) — Dev setup, Docker basics
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) — Common issues
- [.env.example](.env.example) — All config vars
