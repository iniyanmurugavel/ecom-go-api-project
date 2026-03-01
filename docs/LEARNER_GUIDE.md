# Learner's Guide: Understanding This Go API

This guide is for **Go beginners** who want to learn from this project. It explains key concepts, file roles, and how to study the codebase.

---

## 1. Project Structure (What Lives Where)

```
ecom-go-api-project/
├── cmd/                    # Application entry point
│   ├── main.go             # Starts server, loads config, connects to DB
│   └── api.go              # HTTP router, middleware, route wiring
├── internal/               # Private packages (not importable from outside)
│   ├── auth/               # Registration, login, JWT, middleware
│   ├── products/           # Product listing (handler + service + repo interface)
│   ├── orders/             # Order placement (handler + service + repo interface)
│   ├── json/               # JSON read/write helpers
│   ├── validate/           # Input validation (email, password, quantity, etc.)
│   ├── metrics/            # Prometheus metrics
│   ├── env/                # Environment variable helpers
│   └── adapters/postgresql/
│       ├── migrations/     # SQL migrations (goose)
│       └── sqlc/           # Generated Go code from SQL (DO NOT EDIT)
├── docker-compose.yaml     # PostgreSQL container
├── sqlc.yaml               # SQLC config (schema + queries location)
└── .env                    # Secrets (copy from .env.example)
```

**LEARNING:** In Go, `internal/` is special — packages under it can only be imported by the parent module. This keeps your API code private.

---

## 2. Request Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│  CLIENT (Postman, curl, mobile app, browser)                                       │
└────────────────────────────────────────────┬────────────────────────────────────┘
                                              │ HTTP Request
                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  MIDDLEWARE (runs first, in order)                                                │
│  RequestID → RealIP → Logger → Recoverer → Timeout → Metrics → RateLimit → CORS  │
└────────────────────────────────────────────┬────────────────────────────────────┘
                                              │
                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  ROUTER (Chi) — matches URL + method to handler                                   │
│  GET /health, GET /health/live, GET /metrics                                      │
│  POST /v1/auth/register, POST /v1/auth/login                                     │
│  GET /v1/products, POST /v1/orders (require JWT)                                 │
└────────────────────────────────────────────┬────────────────────────────────────┘
                                              │
                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  HANDLER (HTTP layer)                                                             │
│  • Parse JSON body or query params                                                │
│  • Validate input (validate package)                                              │
│  • Call service                                                                   │
│  • Map errors → HTTP status (400, 404, 409, 500)                                  │
│  • Write JSON response                                                           │
└────────────────────────────────────────────┬────────────────────────────────────┘
                                              │
                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  SERVICE (business logic)                                                        │
│  • Validate business rules                                                       │
│  • Start transaction (for orders)                                                │
│  • Call repository                                                              │
│  • Return domain model or error                                                  │
└────────────────────────────────────────────┬────────────────────────────────────┘
                                              │
                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  REPOSITORY (data access) — implements interface                                 │
│  • Production: *repo.Queries (sqlc-generated)                                    │
│  • Tests: mock struct that returns fixed data                                    │
│  • Runs SQL via generated methods                                                │
└────────────────────────────────────────────┬────────────────────────────────────┘
                                              │
                                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│  PostgreSQL (Docker container, port 15432)                                        │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Key Go Concepts Used in This Project

| Concept | Where to See It | What It Means |
|---------|-----------------|---------------|
| **Interfaces** | `orders/repository.go`, `products/repository.go`, `auth/repository.go` | A type that defines methods. Any struct implementing those methods satisfies the interface. Enables mocking in tests. |
| **Context** | `r.Context()`, `auth.WithCustomerID`, `auth.CustomerIDFromContext` | Carries request-scoped values (e.g. customer ID from JWT). Passed through the call chain. |
| **Errors** | `errors.New`, `errors.Is`, `fmt.Errorf("%w", err)` | Sentinel errors for domain (ErrProductNotFound). `%w` wraps errors; `errors.Is` unwraps. |
| **defer** | `defer tx.Rollback(ctx)` in orders/service.go | Runs when the function returns. Used for cleanup (close file, rollback tx). |
| **Goroutines** | `go func() { srv.ListenAndServe() }()` in main.go | Lightweight concurrency. `go` starts a new goroutine. |
| **Channels** | `quit := make(chan os.Signal, 1)`, `<-quit` | Pass data between goroutines. `<-quit` blocks until a value is sent. |
| **Struct tags** | `json:"productId"` in orderItem | Metadata for encoding/decoding. `encoding/json` uses these. |

---

## 4. Layered Architecture (Why We Split It This Way)

| Layer | Responsibility | Example |
|-------|-----------------|---------|
| **Handler** | HTTP only: parse body, status codes, JSON | `orders/handlers.go` reads body, calls service, writes 201 or 400 |
| **Service** | Business logic: validation, transactions | `orders/service.go` validates, runs in tx, decrements stock |
| **Repository** | Data access: SQL via generated code | `sqlc/queries.sql.go` has `CreateOrder`, `FindProductByID` |

**LEARNING:** Handlers don't know about SQL. Services don't know about HTTP. This separation makes testing easy (mock the repo) and keeps each layer focused.

---

## 5. How to Study the Code (Suggested Order)

1. **`cmd/main.go`** — How does the app start? What is `defer`? What is a goroutine?
2. **`cmd/api.go`** — How are routes registered? What is middleware?
3. **`internal/products/handlers.go`** — How does a handler call a service?
4. **`internal/products/service.go`** — How does the service call the repo?
5. **`internal/orders/service.go`** — Why do we use a transaction? What is `defer tx.Rollback`?
6. **`internal/auth/middleware.go`** — How does JWT auth work? What is context?
7. **`internal/adapters/postgresql/sqlc/queries.sql`** — What SQL do we run?
8. **`internal/adapters/postgresql/sqlc/queries.sql.go`** — What did sqlc generate? (Don't edit this file.)

---

## 6. Comments in the Code

Look for comments starting with **LEARNING:** — they explain Go concepts or design decisions for beginners. We keep existing comments and add new ones where they help understanding.

---

## 7. Testing

- **Unit tests** (`*_test.go` in internal/): Use mock repositories. No real DB. Run: `go test ./internal/...`
- **Integration tests** (`cmd/integration_test.go`): Hit real HTTP endpoints. Need DB. Run: `go test ./cmd/...`

---

## 8. Further Reading

- [Go by Example](https://gobyexample.com/) — Short examples of Go syntax
- [Effective Go](https://go.dev/doc/effective_go) — Idiomatic Go style
- [SQLC docs](https://docs.sqlc.dev/) — How SQL → Go code generation works
