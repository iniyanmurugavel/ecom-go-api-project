# Redis Guide — Why and How to Use It in This Project

This guide explains **why Redis is useful**, **when you need it**, and **how to use it** with this ecommerce API. Written for learning.

---

## 1. What Is Redis?

**Redis** = Remote Dictionary Server. It's an **in-memory** key-value store. Data lives in RAM, so reads/writes are very fast (microseconds). It persists to disk optionally.

| Feature | PostgreSQL | Redis |
|---------|------------|-------|
| **Storage** | Disk (with memory cache) | In-memory (optionally persisted) |
| **Data model** | Tables, rows, SQL | Keys, values (strings, hashes, lists, sets) |
| **Use case** | Primary data (products, orders, users) | Caching, sessions, rate limits, pub/sub |
| **Speed** | ms | µs (microseconds) |

**Rule of thumb:** PostgreSQL = source of truth. Redis = fast helper for temporary or derived data.

---

## 2. Why Would This Project Need Redis?

This project **does not use Redis today**. You add it when you need one of these:

| Use case | Why Redis? | Current state |
|----------|------------|---------------|
| **Cache product list** | Products rarely change. Avoid hitting DB on every GET /products. | Every request hits PostgreSQL |
| **Rate limiting** | Count requests per IP in a sliding window. Redis is ideal (fast, TTL). | In-memory per process (lost on restart) |
| **JWT blocklist** | Store revoked token IDs until expiry. Check on every request. | No revocation; tokens valid until expiry |
| **Session storage** | Store session data (e.g. cart) by session ID. | Stateless JWT only |
| **Refresh token store** | Store refresh tokens for revocation on logout. | No refresh tokens yet |

---

## 3. Use Case 1: Caching Product List

**Problem:** `GET /v1/products` hits PostgreSQL on every request. If you have 1000 req/s, that's 1000 DB queries.

**Solution:** Cache the product list in Redis. Key = `products:list`, value = JSON. TTL = 60 seconds (or 5 min if products change rarely).

```
Request 1 → Redis miss → PostgreSQL → store in Redis → return
Request 2..N → Redis hit → return (no DB)
After 60s → cache expires → next request hits DB again
```

**When to use:** High traffic, product list changes infrequently.

**When not to use:** Low traffic (DB is fine), or products change every second (cache would be stale).

---

## 4. Use Case 2: Rate Limiting (Production)

**Problem:** The app uses `httprate.LimitByIP` — in-memory. If you run 3 API instances behind a load balancer, each instance has its own counter. An attacker gets 100 req/min **per instance** = 300 total.

**Solution:** Store rate limit counters in Redis. Key = `ratelimit:{ip}`, value = count. Increment on each request, set TTL = 1 minute. All instances share the same Redis.

**When to use:** Multiple API instances, or you want limits to survive restarts.

---

## 5. Use Case 3: JWT Blocklist (Token Revocation)

**Problem:** JWT is stateless. Once issued, it's valid until expiry. You can't "revoke" it without checking somewhere.

**Solution:** On logout (or password change), add the token's `jti` (JWT ID) to Redis with TTL = token expiry. On every request, check: if `jti` is in Redis → reject (revoked).

```
Key:   blocklist:jti:{jti}
Value: 1 (or empty)
TTL:   same as token expiry (e.g. 24h)
```

**When to use:** You need logout to invalidate tokens immediately.

---

## 6. How to Add Redis to This Project

### Step 1: Run Redis (Docker)

Add to `docker-compose.yaml`:

```yaml
services:
  postgres:
    # ... existing ...

  redis:
    image: redis:7-alpine
    container_name: ecom-redis
    ports:
      - "6379:6379"
    # Optional: persist data
    # volumes:
    #   - redis-data:/data
```

Start: `docker compose up -d`

### Step 2: Go client

```bash
go get github.com/redis/go-redis/v9
```

### Step 3: Connect

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
```

### Step 4: Example — cache products

```go
// Pseudocode
key := "products:list"
cached, err := rdb.Get(ctx, key).Result()
if err == nil {
    // Cache hit — return cached JSON
    return cached
}
// Cache miss — query DB, store in Redis
products := db.ListProducts(ctx)
json := marshal(products)
rdb.Set(ctx, key, json, 60*time.Second)
return json
```

---

## 7. When You Don't Need Redis

| Scenario | Use Redis? |
|----------|------------|
| Learning project, low traffic | No |
| Single instance, < 100 req/min | No |
| PaaS with built-in rate limit | Maybe not |
| Multiple instances, high traffic | Yes (caching, rate limit) |
| Need token revocation | Yes (blocklist) |

---

## 8. Summary

- **Redis** = fast in-memory store for caching, rate limits, sessions, blocklists.
- **This project** = doesn't use Redis yet. PostgreSQL is enough for current scale.
- **Add Redis when:** You need caching, shared rate limits, or JWT revocation across instances.

---

## 9. Related Docs

- [docs/JWT_IMPROVEMENTS.md](JWT_IMPROVEMENTS.md) — JWT blocklist option
- [DEPLOYMENT.md](../DEPLOYMENT.md) — Production setup
