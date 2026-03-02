# Learning Path: Stack-Ranked Order to Master This Project

This document gives you a **step-by-step learning order** for this e-commerce API. Follow it one by one. Each step builds on the previous. All learning docs live in the `docs/` folder.

---

## Overview: What You'll Learn

| # | Topic | Doc / Location | Est. time |
|---|-------|-----------------|-----------|
| 1 | Project setup & run | [HOW_TO_RUN.md](../HOW_TO_RUN.md) | 15 min |
| 2 | CORS: Web vs Mobile | [CORS_AND_CLIENTS.md](CORS_AND_CLIENTS.md) | 10 min |
| 3 | Project structure & request flow | [LEARNER_GUIDE.md](LEARNER_GUIDE.md) | 20 min |
| 4 | Go basics used here | [LEARNER_GUIDE.md](LEARNER_GUIDE.md) § Key Concepts | 15 min |
| 5 | Entry point: main.go | `cmd/main.go` | 15 min |
| 6 | HTTP layer: api.go | `cmd/api.go` | 15 min |
| 7 | Products: handler → service → repo | `internal/products/` | 20 min |
| 8 | Orders: transactions & stock | `internal/orders/` | 25 min |
| 9 | Auth: JWT, middleware, context | `internal/auth/` | 25 min |
| 10 | Validation & JSON helpers | `internal/validate/`, `internal/json/` | 10 min |
| 11 | Database: migrations & SQLC | `internal/adapters/postgresql/` | 25 min |
| 12 | Testing | `*_test.go` files | 20 min |
| 13 | Scalability & production | [REVIEW_AND_LEARNING_ROADMAP.md](../REVIEW_AND_LEARNING_ROADMAP.md) | 20 min |

---

## Step 1: Run the Project (15 min)

**Goal:** Get the API running and hit an endpoint.

1. Read [HOW_TO_RUN.md](../HOW_TO_RUN.md)
2. Run: `docker compose up -d` → `source .env && goose up` → `go run ./cmd`
3. Test: `curl http://localhost:8080/health` → should return `all good`
4. Register and login (see HOW_TO_RUN or AUTH_README), then call `/v1/products` with the token

**Check:** You can start the API and get a 200 response.

---

## Step 2: Understand CORS (10 min)

**Goal:** Know when you need CORS and when you don't.

1. Read [docs/CORS_AND_CLIENTS.md](CORS_AND_CLIENTS.md)
2. Key takeaway: CORS is for **browsers** only. Mobile apps and Postman don't need it.
3. If you'll build a React frontend: add `CORS_ALLOWED_ORIGINS=http://localhost:3000` to `.env`

**Check:** You can explain why a React app at localhost:3000 needs CORS but a mobile app doesn't.

---

## Step 3: Project Structure & Request Flow (20 min)

**Goal:** See the big picture before diving into code.

1. Read [docs/LEARNER_GUIDE.md](LEARNER_GUIDE.md) — sections 1 and 2
2. Study the **Request Flow Diagram** (Client → Middleware → Router → Handler → Service → Repo → DB)
3. Glance at the folder structure; know where `cmd/`, `internal/`, `migrations/`, `sqlc/` live

**Check:** You can draw the request flow from memory.

---

## Step 4: Go Concepts Used Here (15 min)

**Goal:** Understand interfaces, context, errors, defer, goroutines, channels.

1. Read [docs/LEARNER_GUIDE.md](LEARNER_GUIDE.md) — section 3 (Key Go Concepts)
2. Keep this as a reference; you'll see these in the code

**Check:** You know what an interface is and why we use `defer tx.Rollback`.

---

## Step 5: Entry Point — main.go (15 min)

**Goal:** Understand how the app starts.

1. Open `cmd/main.go`
2. Follow the flow: load .env → create logger → parse config → connect to DB (pool) → mount router → start server (goroutine) → wait for shutdown signal
3. Read every **LEARNING:** comment in the file

**Check:** You can explain what `defer pool.Close()` and `<-quit` do.

---

## Step 6: HTTP Layer — api.go (15 min)

**Goal:** Understand middleware, routes, and how handlers are wired.

1. Open `cmd/api.go`
2. See middleware order: RequestID, RealIP, Logger, Recoverer, Timeout, Metrics, RateLimit, CORS
3. See route groups: `/health`, `/metrics`, `/v1` (auth + protected)
4. See how `auth.RequireAuth` wraps protected routes

**Check:** You can explain what runs before a handler and why order matters.

---

## Step 7: Products — Handler → Service → Repo (20 min)

**Goal:** Understand the layered flow for a simple read-only endpoint.

1. Read `internal/products/handlers.go` — how it reads `limit`/`offset`, calls service, writes JSON
2. Read `internal/products/service.go` — how it calls the repo
3. Read `internal/products/repository.go` — the interface
4. See `internal/adapters/postgresql/sqlc/queries.sql` — `ListProductsPaginated`
5. See `internal/adapters/postgresql/sqlc/queries.sql.go` — generated code (don't edit)

**Check:** You can trace a `GET /v1/products` request from Chi to the database and back.

---

## Step 8: Orders — Transactions & Stock (25 min)

**Goal:** Understand transactions, rollback, and stock decrement.

1. Read `internal/orders/handlers.go` — how customer ID comes from JWT context
2. Read `internal/orders/service.go` — transaction, `defer tx.Rollback`, loop over items, `DecrementProductStock`
3. Read `internal/orders/repository.go` — `OrderTxRepo`, `WithTx`
4. Read `internal/orders/types.go` — `createOrderParams`, `orderItem`

**Check:** You can explain why we use a transaction and what happens if `DecrementProductStock` fails.

---

## Step 9: Auth — JWT, Middleware, Context (25 min)

**Goal:** Understand registration, login, JWT, and how customer ID flows to handlers.

1. Read `internal/auth/handlers.go` — Register, Login
2. Read `internal/auth/service.go` — bcrypt, CreateCustomer, FindCustomerByEmail
3. Read `internal/auth/jwt.go` — SignToken, VerifyToken
4. Read `internal/auth/middleware.go` — RequireAuth, how it puts customer ID in context
5. Read `internal/auth/context.go` — WithCustomerID, CustomerIDFromContext

**Check:** You can explain the full flow: login → token → Authorization header → middleware → context → handler.

---

## Step 10: Validation & JSON (10 min)

**Goal:** Know where validation happens and how errors are formatted.

1. Read `internal/validate/validate.go` — Email, Password, Name, OrderItemQuantity, ProductID
2. Read `internal/json/json.go` — Write, WriteError (with request_id), Read

**Check:** You know that handlers validate before calling services, and errors use `json.WriteError`.

---

## Step 11: Database — Migrations & SQLC (25 min)

**Goal:** Understand schema, migrations, and how SQL becomes Go code.

1. Read `internal/adapters/postgresql/migrations/00001_create_products.sql`
2. Read `internal/adapters/postgresql/migrations/00002_create_orders.sql`
3. Read `internal/adapters/postgresql/migrations/00003_create_customers.sql`
4. Read `internal/adapters/postgresql/sqlc/queries.sql` — all queries
5. Read [ARCHITECTURE.md](../ARCHITECTURE.md) — section 3 (SQLC) and section 5 (when to run migrations, sqlc generate)

**Check:** You can add a new migration and a new query, then run `goose up` and `sqlc generate`.

---

## Step 12: Testing (20 min)

**Goal:** Understand unit tests (mocks) and integration tests (real HTTP + DB).

1. Read `internal/products/service_test.go` — mock repository
2. Read `internal/orders/service_test.go` — mock OrderTxRepo, DecrementProductStock
3. Read `internal/auth/service_test.go` — mock AuthRepository
4. Read `cmd/integration_test.go` — real HTTP calls
5. Run: `go test ./internal/...` and `go test ./cmd/...`

**Check:** You can run tests and explain why we use mocks for services.

---

## Step 13: Scalability & Production (20 min)

**Goal:** Know what this project does well and what to add for production.

1. Read [REVIEW_AND_LEARNING_ROADMAP.md](../REVIEW_AND_LEARNING_ROADMAP.md)
2. Note: connection pool, interfaces, graceful shutdown, health, rate limit, metrics, CORS

**Check:** You can list 3 things this API does for scalability and 2 things you'd add for production.

---

## Docs Folder Reference

All learning docs in one place:

| File | Purpose |
|------|---------|
| [docs/LEARNING_PATH.md](LEARNING_PATH.md) | **Start here** — this stack-ranked order |
| [docs/CORS_AND_CLIENTS.md](CORS_AND_CLIENTS.md) | CORS explained: web vs mobile, with examples |
| [docs/LEARNER_GUIDE.md](LEARNER_GUIDE.md) | Project structure, request flow, Go concepts |
| [docs/MAC_POSTGRES_15432.md](MAC_POSTGRES_15432.md) | Use Mac Postgres on 15432 instead of Docker |

**Root docs** (project root):

| File | Purpose |
|------|---------|
| [HOW_TO_RUN.md](../HOW_TO_RUN.md) | Run the project |
| [ARCHITECTURE.md](../ARCHITECTURE.md) | High-level design, SQLC, migrations |
| [AUTH_README.md](../AUTH_README.md) | JWT auth flow |
| [TROUBLESHOOTING.md](../TROUBLESHOOTING.md) | Common issues |
| [REVIEW_AND_LEARNING_ROADMAP.md](../REVIEW_AND_LEARNING_ROADMAP.md) | Scalability & next steps |
