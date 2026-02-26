# Authentication (JWT) — How It Works

This document explains how JWT authentication works in this API, how to test it in Postman, and how customer data flows from the token into your requests.

---

## 1. Overview

- **All product and order APIs require a JWT** in the `Authorization` header.
- **Auth endpoints** (`/v1/auth/register`, `/v1/auth/login`) are **public** — no token needed.
- **Flow:** Register or login → get a JWT → use it in `Authorization: Bearer <token>` for `/v1/products` and `/v1/orders`.
- **Customer ID** is stored inside the JWT. When you place an order, the API uses the customer ID from the token (not from the request body).

---

## 2. How JWT Works (Simple Explanation)

1. **Register** or **Login** with email and password.
2. The server verifies your credentials and creates a **JWT** (JSON Web Token).
3. The JWT contains your **customer ID** and **email** (and expiry time).
4. You send the JWT in every request to protected endpoints.
5. The server reads the JWT, extracts your customer ID, and uses it (e.g. for orders).

**Why JWT?** The server does not need to store sessions. The token is self-contained and can be verified using a secret. Stateless and scalable.

---

## 3. Auth Endpoints

### POST /v1/auth/register

Creates a new customer. No token required.

**Request:**
```json
{
  "email": "john@example.com",
  "password": "secret123",
  "name": "John Doe"
}
```

**Response (201 Created):**
```json
{
  "id": 1,
  "email": "john@example.com",
  "name": "John Doe"
}
```

**Errors:**
- `400` — Invalid JSON or missing email/password/name
- `409` — Email already registered

---

### POST /v1/auth/login

Logs in and returns a JWT. No token required.

**Request:**
```json
{
  "email": "john@example.com",
  "password": "secret123"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "customer_id": 1,
  "email": "john@example.com"
}
```

**Errors:**
- `400` — Invalid JSON or missing email/password
- `401` — Invalid email or password

**Use the `token`** in the `Authorization` header for all protected requests.

---

## 4. Protected Endpoints (Require JWT)

### GET /v1/products

List products with pagination. Requires `Authorization: Bearer <token>`.

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Query params (optional):** `?limit=20&offset=0`

---

### POST /v1/orders

Place an order. **Customer ID comes from the JWT** — you do **not** send `customerId` in the body.

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

**Request body:**
```json
{
  "items": [
    { "productId": 1, "quantity": 2 },
    { "productId": 2, "quantity": 1 }
  ]
}
```

**Response (201 Created):**
```json
{
  "id": 10,
  "customer_id": 1,
  "created_at": "2026-02-26T21:12:13.573892+05:30"
}
```

**Errors:**
- `401` — Missing or invalid token
- `400` — Invalid JSON or empty items
- `404` — Product not found
- `409` — Not enough stock

---

## 5. Postman: Step-by-Step Testing

### Step 1: Register a customer

1. Create a new request.
2. **Method:** `POST`
3. **URL:** `http://localhost:8080/v1/auth/register`
4. **Headers:** `Content-Type: application/json`
5. **Body (raw, JSON):**
   ```json
   {
     "email": "test@example.com",
     "password": "password123",
     "name": "Test User"
   }
   ```
6. Click **Send**. You should get `201` and `{"id": 1, "email": "test@example.com", "name": "Test User"}`.

---

### Step 2: Login to get a token

1. Create a new request.
2. **Method:** `POST`
3. **URL:** `http://localhost:8080/v1/auth/login`
4. **Headers:** `Content-Type: application/json`
5. **Body (raw, JSON):**
   ```json
   {
     "email": "test@example.com",
     "password": "password123"
   }
   ```
6. Click **Send**. Copy the `token` from the response.

---

### Step 3: Use the token for products

1. Create a new request.
2. **Method:** `GET`
3. **URL:** `http://localhost:8080/v1/products?limit=10&offset=0`
4. **Headers:** `Authorization: Bearer <paste-your-token-here>`
5. Click **Send**. You should get `200` and a JSON array of products.

---

### Step 4: Use the token for orders

1. Create a new request.
2. **Method:** `POST`
3. **URL:** `http://localhost:8080/v1/orders`
4. **Headers:**
   - `Authorization: Bearer <paste-your-token-here>`
   - `Content-Type: application/json`
5. **Body (raw, JSON):**
   ```json
   {
     "items": [
       { "productId": 1, "quantity": 2 },
       { "productId": 2, "quantity": 1 }
     ]
   }
   ```
6. Click **Send**. You should get `201` and the created order.

---

### Postman: Save token in an environment variable (optional)

1. In the **Login** request, go to the **Tests** tab.
2. Add:
   ```javascript
   var json = pm.response.json();
   pm.environment.set("token", json.token);
   ```
3. In **Products** and **Orders** requests, set the header:
   ```
   Authorization: Bearer {{token}}
   ```
4. After login, `{{token}}` will be replaced with the actual token.

---

## 6. How the Code Works

| Component | File | Role |
|-----------|------|------|
| **JWT sign/verify** | `internal/auth/jwt.go` | Creates and validates tokens; claims include `customer_id` and `email` |
| **Context** | `internal/auth/context.go` | `CustomerIDFromContext(ctx)` gets the customer ID; `WithCustomerID` sets it |
| **Middleware** | `internal/auth/middleware.go` | `RequireAuth` reads `Authorization: Bearer <token>`, verifies JWT, puts customer ID in context |
| **Auth handlers** | `internal/auth/handlers.go` | Register (hash password, create customer); Login (verify password, return JWT) |
| **Orders handler** | `internal/orders/handlers.go` | Gets `customerID` from `auth.CustomerIDFromContext(r.Context())`; body has only `items` |

**Flow for POST /v1/orders:**
1. Request has `Authorization: Bearer <token>`.
2. `RequireAuth` middleware verifies the token and calls `WithCustomerID(ctx, claims.CustomerID)`.
3. Orders handler calls `auth.CustomerIDFromContext(r.Context())` to get the customer ID.
4. Handler builds `createOrderParams{CustomerID: customerID, Items: body.Items}` and calls the service.

---

## 7. Configuration

Add to `.env`:

```
JWT_SECRET=your-long-secret-at-least-32-characters
```

- **Development:** Default is `change-me-in-production-use-long-secret` (you'll see a warning).
- **Production:** Use a long, random secret (e.g. 32+ chars). Never commit it.

---

## 8. Token Expiry

Tokens expire after **24 hours** (see `auth.DefaultExpiry`). After expiry, the client gets `401` and must login again to get a new token.

---

## 9. Quick Reference

| Endpoint | Method | Auth | Body |
|----------|--------|------|------|
| `/v1/auth/register` | POST | No | `{ email, password, name }` |
| `/v1/auth/login` | POST | No | `{ email, password }` |
| `/v1/products` | GET | Bearer token | (none) |
| `/v1/orders` | POST | Bearer token | `{ items: [{ productId, quantity }] }` |
