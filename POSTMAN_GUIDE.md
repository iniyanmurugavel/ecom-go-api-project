# Postman Guide — How to Call the Ecommerce API

This document explains how to install Postman, set up requests, and call every API endpoint. Use it as a standalone reference for testing the API.

---

## 1. Install Postman

### Option A: Postman Desktop App (recommended for manual testing)

1. Download from [postman.com/downloads](https://www.postman.com/downloads/)
2. Install and open Postman
3. Create a free account (optional) or use without signing in

### Option B: Postman CLI (for automation / CI)

Install the Postman CLI globally via npm:

```bash
npm install -g postman-cli
```

Verify installation:

```bash
postman --version
```

**Use Postman CLI when:**
- Running API tests from the terminal
- Automating requests in scripts
- Running Newman (Postman's CLI runner) for collection runs

**Use Postman Desktop when:**
- Manually exploring and testing endpoints
- Debugging requests and responses
- Saving collections and environments

---

## 2. Prerequisites

Before calling the API in Postman:

1. **Start the API** — Run `./scripts/setup-and-run.sh` from the project root (or `go run ./cmd` after DB is up).
2. **Base URL** — All requests use: `http://localhost:8080`
3. **Protected endpoints** — Products and Orders require a JWT. Get it via Login first.

---

## 3. API Endpoints — Step by Step

### 3.1 Health Check (no auth)

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `http://localhost:8080/health` |
| **Headers** | (none) |
| **Body** | (none) |

**Expected response:** `200 OK` — Body: `all good`

---

### 3.2 Health with DB Check (no auth)

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `http://localhost:8080/health/live` |
| **Headers** | (none) |
| **Body** | (none) |

**Expected response:** `200 OK` if DB is up; `503` if DB is down.

---

### 3.3 Register (no auth)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `http://localhost:8080/v1/auth/register` |
| **Headers** | `Content-Type: application/json` |
| **Body** | Raw → JSON |

**Body (raw JSON):**

```json
{
  "email": "you@example.com",
  "password": "secret123",
  "name": "Your Name"
}
```

**Expected response:** `201 Created`

```json
{
  "id": 1,
  "email": "you@example.com",
  "name": "Your Name"
}
```

**Errors:**
- `400` — Invalid JSON or missing email/password/name
- `409` — Email already registered

---

### 3.4 Login (no auth)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `http://localhost:8080/v1/auth/login` |
| **Headers** | `Content-Type: application/json` |
| **Body** | Raw → JSON |

**Body (raw JSON):**

```json
{
  "email": "you@example.com",
  "password": "secret123"
}
```

**Expected response:** `200 OK`

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "customer_id": 1,
  "email": "you@example.com"
}
```

**Important:** Copy the `token` value. Use it in the `Authorization` header for Products and Orders.

**Errors:**
- `400` — Invalid JSON or missing email/password
- `401` — Invalid email or password

---

### 3.5 List Products (requires JWT)

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `http://localhost:8080/v1/products?limit=20&offset=0` |
| **Headers** | `Authorization: Bearer <your-token>` |
| **Body** | (none) |

**Query params (optional):**
- `limit` — Max products to return (default: 20)
- `offset` — Skip first N products (for pagination)

**Expected response:** `200 OK`

```json
[
  {
    "id": 1,
    "name": "Sample Product A",
    "price_in_centers": 1999,
    "quantity": 10,
    "created_at": "2026-02-26T22:24:51.860891+05:30"
  },
  ...
]
```

**Errors:**
- `401` — Missing or invalid token

---

### 3.6 Place Order (requires JWT)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `http://localhost:8080/v1/orders` |
| **Headers** | `Authorization: Bearer <your-token>`<br>`Content-Type: application/json` |
| **Body** | Raw → JSON |

**Body (raw JSON):** Customer ID comes from the JWT — do **not** send `customerId` in the body.

```json
{
  "items": [
    { "productId": 1, "quantity": 2 },
    { "productId": 2, "quantity": 1 }
  ]
}
```

Use real `productId` values from the **List products** response.

**Expected response:** `201 Created`

```json
{
  "id": 10,
  "customer_id": 1,
  "created_at": "2026-02-26T22:12:13.573892+05:30"
}
```

**Errors:**
- `401` — Missing or invalid token
- `400` — Invalid JSON or empty items
- `404` — Product not found
- `409` — Not enough stock

---

## 4. Quick Reference Table

| Endpoint | Method | Auth | Headers | Body |
|----------|--------|------|---------|------|
| `/health` | GET | No | — | — |
| `/health/live` | GET | No | — | — |
| `/v1/auth/register` | POST | No | `Content-Type: application/json` | `{ email, password, name }` |
| `/v1/auth/login` | POST | No | `Content-Type: application/json` | `{ email, password }` |
| `/v1/products` | GET | Bearer token | `Authorization: Bearer <token>` | — |
| `/v1/orders` | POST | Bearer token | `Authorization: Bearer <token>`, `Content-Type: application/json` | `{ items: [{ productId, quantity }] }` |

---

## 5. Save Token in Postman (Optional)

To avoid copying the token manually for every request:

1. In the **Login** request, open the **Tests** tab.
2. Add this script:

   ```javascript
   var json = pm.response.json();
   pm.environment.set("token", json.token);
   ```

3. Create an environment: **Environments** → **Create Environment** → name it `Local`.
4. Select the `Local` environment (top-right dropdown).
5. In **Products** and **Orders** requests, set the header:
   - Key: `Authorization`
   - Value: `Bearer {{token}}`

After you run **Login**, the `{{token}}` variable will be set automatically. All subsequent requests will use it.

---

## 6. Suggested Order of Testing

1. **Health** — Verify API is running.
2. **Register** — Create a user.
3. **Login** — Get the token.
4. **Products** — List products (use token).
5. **Orders** — Place an order (use token; use `productId` from products).

---

## 7. Postman CLI (Newman) — Run a Collection

If you export a Postman collection and want to run it from the terminal:

```bash
# Install Newman (Postman's CLI runner)
npm install -g newman

# Run a collection (replace with your exported file)
newman run collection.json -e environment.json
```

For more on Postman CLI and Newman, see [Postman CLI docs](https://learning.postman.com/docs/postman-cli/postman-cli-overview/).

---

## 8. Related Docs

- **[AUTH_README.md](AUTH_README.md)** — How JWT auth works, token expiry, config
- **[README.md](README.md)** — Project setup, curl examples, architecture
