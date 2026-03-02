# Code Review & Learning Roadmap

This document is a **scalability and architecture review** of your e-commerce API, plus a **learning roadmap** for building production-ready APIs. Use it as a reference to improve this project and future ones.

**New to Go?** Start with [docs/LEARNER_GUIDE.md](docs/LEARNER_GUIDE.md) for key concepts and a study path. The code has **LEARNING:** comments for beginners.

---

## 1. Overall Assessment

| Aspect | Rating | Notes |
|--------|--------|-------|
| **Architecture** | ⭐⭐⭐⭐ | Clean Handler → Service → Repo layering; interfaces for testability |
| **Scalability** | ⭐⭐⭐ | Connection pool, stateless design; missing pool tuning, metrics |
| **Security** | ⭐⭐⭐ | JWT, bcrypt, rate limiting; needs input validation, CORS |
| **Production readiness** | ⭐⭐ | Missing observability, validation, one critical bug (stock) |
| **Testability** | ⭐⭐⭐⭐ | Unit tests with mocks, integration tests, CI |

**Verdict:** Strong learning foundation. The patterns (layering, interfaces, SQLC, migrations) are production-grade. A few critical fixes and incremental improvements will make it a solid reference for real APIs.

---

## 2. What You Did Well

### Architecture
- **Layered design**: Handler (HTTP) → Service (logic) → Repo (data). Clear separation of concerns.
- **Repository interfaces**: `products.Repository`, `orders.OrderRepo`, `orders.TxBeginner` let you mock in tests without a real DB.
- **SQLC**: Type-safe SQL, no raw strings in Go. Schema and queries live in SQL files.
- **Transactions**: Order placement runs in one transaction; rollback on any error.

### Scalability
- **Connection pool** (`pgxpool`): Many concurrent requests share connections instead of one-at-a-time.
- **Stateless API**: No server-side session; JWT carries identity. Easy to scale horizontally.
- **Pagination**: Products list uses `limit`/`offset`; responses stay bounded.
- **Graceful shutdown**: Waits for in-flight requests before exit (deploy-friendly).

### Operations
- **Health endpoints**: `/health` (simple) and `/health/live` (DB ping for load balancers).
- **Rate limiting**: Per-IP limit protects DB and API.
- **CI**: Migrations, build, unit + integration tests on push/PR.

### Security
- **JWT auth**: Bearer token; customer ID from context (not body).
- **bcrypt**: Password hashing with default cost.
- **Unique constraint**: Email uniqueness enforced at DB level.

---

## 3. Critical Issues (Fix First)

### 3.1 Stock Not Decremented (Bug)

**Location:** `internal/orders/service.go` line 78

When an order is placed, product stock is **never decremented**. Orders succeed but inventory stays the same.

**Fix:** Add a query and call it inside the transaction:

```sql
-- In queries.sql
-- name: DecrementProductStock :exec
UPDATE products SET quantity = quantity - $1 WHERE id = $2 AND quantity >= $1;
```

Then in the service, after `CreateOrderItem`, call `DecrementProductStock`. If `rows affected == 0`, return `ErrProductNoStock` (handles race: another order took the last units).

### 3.2 Order Items: No FK to Products

**Location:** `internal/adapters/postgresql/migrations/00002_create_orders.sql`

`order_items.product_id` has no foreign key to `products(id)`. You can insert `product_id = 99999` even if that product doesn’t exist.

**Fix:** Add migration:

```sql
ALTER TABLE order_items ADD CONSTRAINT fk_product FOREIGN KEY (product_id) REFERENCES products(id);
```

(Your service checks `FindProductByID` first, so this is a safety net, not a logic fix.)

### 3.3 Inconsistent Error Response Format

Some handlers use `http.Error(w, err.Error(), 400)` (plain text), others use `json.Write` (JSON). Clients expect consistent JSON.

**Fix:** Use a shared error response helper, e.g. `json.WriteError(w, status, "error", "message")`, and use it everywhere.

---

## 4. Important Gaps for Production

### 4.1 Input Validation

- **Auth**: No max length on email, password, name. No email format check.
- **Orders**: No check that `quantity > 0` or `productId > 0`.
- **Products pagination**: Negative `limit`/`offset` are handled, but very large values could be problematic.

**Suggestion:** Use `github.com/go-playground/validator/v10` or manual checks. Validate in handlers or a validation middleware.

### 4.2 CORS

If a browser frontend calls this API, you need CORS headers. Without them, browsers block cross-origin requests.

**Suggestion:** Add `github.com/go-chi/cors` and configure allowed origins.

### 4.3 Logging Consistency

- `products` and `orders` use injected `*slog.Logger`.
- `auth/handlers.go` uses `log.Println`.

**Fix:** Inject a logger into auth handlers and use it instead of `log`.

### 4.4 Connection Pool Tuning

`pgxpool.New(ctx, dsn)` uses defaults. For production you often want:

- `MaxConns`: cap concurrent DB connections.
- `MinConns`: keep a minimum pool size.
- `MaxConnLifetime`, `MaxConnIdleTime`: avoid stale connections.

**Suggestion:** Use `pgxpool.Config` and set these from env vars.

### 4.5 Auth Service Testability

Auth service depends on concrete `*repo.Queries`, not an interface. Harder to unit test without a DB.

**Suggestion:** Define `AuthRepository` (e.g. `CreateCustomer`, `FindCustomerByEmail`) and inject it.

---

## 5. Nice-to-Have for Production

| Area | What to Add |
|------|-------------|
| **API docs** | OpenAPI/Swagger (e.g. `go-swagger` or `oapi-codegen`) |
| **Metrics** | Prometheus metrics (request count, latency, errors) |
| **Tracing** | OpenTelemetry for request tracing across services |
| **Retries** | Retry with backoff for transient DB/network errors |
| **Config validation** | Fail fast if required env vars (e.g. `JWT_SECRET`) are missing or weak |
| **Request ID** | Chi’s `RequestID` is present; ensure it’s in logs and error responses |
| **Structured errors** | `{"error": "...", "code": "PRODUCT_NOT_FOUND", "request_id": "..."}` |

---

## 6. Scalability Deep Dive

### What Already Scales

1. **Stateless**: No in-memory session; JWT is self-contained.
2. **Connection pool**: Handles many concurrent requests.
3. **Pagination**: Prevents unbounded responses.
4. **Graceful shutdown**: Avoids dropped requests during deploys.

### What to Add as Traffic Grows

1. **Pool size**: Tune `MaxConns` based on DB and app capacity.
2. **Read replicas**: For read-heavy workloads, route `ListProducts` to a replica.
3. **Caching**: Cache product list (e.g. Redis) with TTL if data is not real-time.
4. **Rate limiting**: You have per-IP; consider per-user or per-API-key for paid tiers.
5. **Horizontal scaling**: Run multiple API instances behind a load balancer; health checks use `/health/live`.

---

## 7. Learning Roadmap: Next Steps

Use this as a checklist to level up from “learning project” to “production reference.”

### Phase 1: Fix Critical (1–2 days) ✅ DONE
- [x] Implement stock decrement in `PlaceOrder` (with `DecrementProductStock` query).
- [x] Add FK from `order_items.product_id` to `products.id` (migration 00004).
- [x] Standardize error responses to JSON (`json.WriteError` with request_id).

### Phase 2: Validation & Consistency (1–2 days) ✅ DONE
- [x] Add input validation (email format, quantity > 0, productId > 0, password length).
- [x] Inject logger into auth handlers; remove `log.Println`.
- [x] Add CORS middleware (configurable via `CORS_ALLOWED_ORIGINS`).

### Phase 3: Observability (2–3 days) ✅ DONE
- [x] Add Prometheus metrics (request count, latency, status codes) at `/metrics`.
- [x] Request ID from chi middleware included in error responses.
- [x] Structured error responses with `request_id`.

### Phase 4: Production Hardening (2–3 days) ✅ PARTIAL
- [x] Configure connection pool from env (`DB_POOL_MAX_CONNS`, `DB_POOL_MIN_CONNS`).
- [x] Add auth repository interface and unit tests.
- [ ] Add OpenAPI spec and generate docs.
- [ ] Add retry logic for DB connection.

### Phase 5: Advanced (Optional)
- [ ] Add refresh tokens for JWT.
- [ ] Add product search/filtering.
- [ ] Add order history by customer.
- [ ] Dockerize the API for deployment.

---

## 8. Summary

Your project shows solid understanding of:

- Layered architecture and dependency injection
- Type-safe SQL with SQLC
- Auth (JWT, bcrypt)
- Testing (mocks, integration tests)
- CI and migrations

The main gaps are:

1. **One critical bug**: stock not decremented on order.
2. **Validation**: input validation and consistent error format.
3. **Observability**: metrics, tracing, structured errors.
4. **Production config**: pool tuning, CORS, stricter env validation.

Tackle Phase 1 first, then Phase 2. After that, this codebase is a strong reference for building production APIs in Go.
