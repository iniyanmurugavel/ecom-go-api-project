# API Reference — All Endpoints with curl and Responses

Single source of truth for every endpoint: method, URL, headers, body, and expected responses. Use for curl, Postman, or any HTTP client.

**Base URL:** `http://localhost:8080`

---

## 1. Health Check

**GET** `/health`

No auth. Quick liveness check.

### curl

```bash
curl http://localhost:8080/health
```

### Response (200 OK)

```
all good
```

---

## 2. Health with DB Check

**GET** `/health/live`

No auth. Pings the database. Returns 503 if DB is down.

### curl

```bash
curl http://localhost:8080/health/live
```

### Response (200 OK)

```
ok
```

### Response (503 Service Unavailable)

```json
{"error":"database unavailable","request_id":"abc123"}
```

---

## 3. Metrics (Prometheus)

**GET** `/metrics`

No auth. Prometheus-format metrics.

### curl

```bash
curl http://localhost:8080/metrics
```

### Response (200 OK)

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/health",status="200"} 5
...
```

---

## 4. Register

**POST** `/v1/auth/register`

No auth. Creates a new customer.

### curl

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"secret123","name":"Your Name"}'
```

### Request body

```json
{
  "email": "you@example.com",
  "password": "secret123",
  "name": "Your Name"
}
```

### Response (201 Created)

```json
{
  "id": 1,
  "email": "you@example.com",
  "name": "Your Name"
}
```

### Error responses

| Status | Body |
|--------|------|
| 400 | `{"error":"field is required"}` or validation message |
| 409 | `{"error":"email already registered"}` |
| 500 | `{"error":"registration failed"}` |

---

## 5. Login

**POST** `/v1/auth/login`

No auth. Returns JWT token.

### curl

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"secret123"}'
```

### Request body

```json
{
  "email": "you@example.com",
  "password": "secret123"
}
```

### Response (200 OK)

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "customer_id": 1,
  "email": "you@example.com"
}
```

**Use the `token` value** in `Authorization: Bearer <token>` for Products and Orders.

### Error responses

| Status | Body |
|--------|------|
| 400 | `{"error":"invalid email format"}` or validation message |
| 401 | `{"error":"invalid email or password"}` |
| 500 | `{"error":"login failed"}` |

---

## 6. List Products

**GET** `/v1/products`

**Requires:** `Authorization: Bearer <token>`

### curl

```bash
curl "http://localhost:8080/v1/products?limit=20&offset=0" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Query params (optional)

| Param | Default | Description |
|-------|---------|-------------|
| limit | 20 | Max products (max 100) |
| offset | 0 | Skip first N (pagination) |

### Response (200 OK)

```json
[
  {
    "id": 1,
    "name": "Sample Product A",
    "price_in_centers": 1999,
    "quantity": 10,
    "created_at": "2026-02-26T22:24:51.860891+05:30"
  },
  {
    "id": 2,
    "name": "Sample Product B",
    "price_in_centers": 2999,
    "quantity": 5,
    "created_at": "2026-02-26T22:24:51.860891+05:30"
  }
]
```

### Error responses

| Status | Body |
|--------|------|
| 401 | `{"error":"missing Authorization header"}` or `{"error":"invalid or expired token"}` |
| 500 | `{"error":"internal server error"}` |

---

## 7. Place Order

**POST** `/v1/orders`

**Requires:** `Authorization: Bearer <token>`

Customer ID comes from the JWT — do **not** send `customerId` in the body.

### curl

```bash
curl -X POST http://localhost:8080/v1/orders \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"productId":1,"quantity":2},{"productId":2,"quantity":1}]}'
```

### Request body

```json
{
  "items": [
    { "productId": 1, "quantity": 2 },
    { "productId": 2, "quantity": 1 }
  ]
}
```

Use real `productId` values from **List Products**.

### Response (201 Created)

```json
{
  "id": 10,
  "customer_id": 1,
  "created_at": "2026-02-26T22:12:13.573892+05:30"
}
```

### Error responses

| Status | Body |
|--------|------|
| 400 | `{"error":"quantity must be greater than 0"}` or validation message |
| 401 | `{"error":"unauthorized"}` |
| 404 | `{"error":"product not found"}` |
| 409 | `{"error":"product has not enough stock"}` |
| 500 | `{"error":"internal server error"}` |

---

## 8. Quick Copy-Paste Flow

```bash
# 1. Health
curl http://localhost:8080/health

# 2. Register
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123","name":"Test User"}'

# 3. Login (save token from response)
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}'

# 4. Products (replace TOKEN)
curl "http://localhost:8080/v1/products?limit=5" \
  -H "Authorization: Bearer TOKEN"

# 5. Order (replace TOKEN)
curl -X POST http://localhost:8080/v1/orders \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"productId":1,"quantity":1}]}'
```

---

## 9. Error Response Format

All JSON errors use this shape:

```json
{
  "error": "human-readable message",
  "request_id": "optional-id-from-chi-middleware"
}
```
