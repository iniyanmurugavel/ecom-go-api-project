# Postman CLI (Newman) — Run API Tests from Terminal

This guide explains how to use **Newman** (Postman's CLI) to run the Ecom API collection from the terminal. Useful for automation, CI, or quick smoke tests.

---

## 1. What You Need

- **Node.js** (for npm)
- **API running** at `http://localhost:8080` (run `go run ./cmd` first)
- **PostgreSQL** running (Docker)

---

## 2. Install Newman

```bash
npm install -g newman
```

Verify:

```bash
newman --version
```

---

## 3. Run the Collection

From the **project root**:

```bash
newman run postman/Ecom-API.postman_collection.json -e postman/Local.postman_environment.json
```

**What this does:**
- `postman/Ecom-API.postman_collection.json` — the collection (Health, Register, Login, Products, Orders)
- `-e postman/Local.postman_environment.json` — environment with `base_url=http://localhost:8080`

---

## 4. Expected Output

```
Ecom API

→ 1. Health Check
  GET http://localhost:8080/health [200 OK, 9B, 45ms]
→ 2. Register
  POST http://localhost:8080/v1/auth/register [201 Created, 52B, 12ms]
→ 3. Login
  POST http://localhost:8080/v1/auth/login [200 OK, 156B, 8ms]
→ 4. Get Products
  GET http://localhost:8080/v1/products?limit=20&offset=0 [200 OK, 234B, 5ms]
→ 5. Place Order
  POST http://localhost:8080/v1/orders [201 Created, 67B, 15ms]

┌─────────────────────────┬────────────────────┬───────────────────┐
│                         │           executed │            failed │
├─────────────────────────┼────────────────────┼───────────────────┤
│              iterations │                  1 │                 0 │
├─────────────────────────┼────────────────────┼───────────────────┤
│                requests │                  5 │                 0 │
├─────────────────────────┼────────────────────┼───────────────────┤
│            test-scripts │                  1 │                 0 │
├─────────────────────────┼────────────────────┼───────────────────┤
│      prerequest-scripts │                  0 │                 0 │
├─────────────────────────┼────────────────────┼───────────────────┤
│              assertions │                  0 │                 0 │
├─────────────────────────┴────────────────────┴───────────────────┤
│ total run duration: 95ms                                          │
└───────────────────────────────────────────────────────────────────┘
```

---

## 5. Run with Different Options

### Different port (API on 8081)

Edit `postman/Local.postman_environment.json` and change `base_url` to `http://localhost:8081`, or use Newman's env override:

```bash
newman run postman/Ecom-API.postman_collection.json \
  -e postman/Local.postman_environment.json \
  --env-var "base_url=http://localhost:8081"
```

### Verbose output

```bash
newman run postman/Ecom-API.postman_collection.json \
  -e postman/Local.postman_environment.json \
  --verbose
```

### HTML report

```bash
npm install -g newman-reporter-htmlextra
newman run postman/Ecom-API.postman_collection.json \
  -e postman/Local.postman_environment.json \
  -r htmlextra
```

Creates `newman/report.html`.

### CI (non-interactive)

```bash
newman run postman/Ecom-API.postman_collection.json \
  -e postman/Local.postman_environment.json \
  --bail
```

`--bail` stops on first failure.

---

## 6. Collection Order

The collection runs in this order:

1. **Health Check** — Verifies API is up
2. **Register** — Creates user (may fail with 409 if already exists; that's OK)
3. **Login** — Gets token; script saves it to `{{token}}`
4. **Get Products** — Uses `{{token}}` in Authorization header
5. **Place Order** — Uses `{{token}}`; customer ID from JWT

**Note:** If Register fails with 409 (email exists), Login will still work. Newman continues.

---

## 7. Troubleshooting

| Issue | Fix |
|-------|-----|
| `Error: connect ECONNREFUSED` | Start the API: `go run ./cmd` |
| `401` on Products/Orders | Login may have failed. Check Register used `test@example.com` / `secret123`; Login uses same. |
| `404` on Place Order | Products table may be empty. Run seed script or add products via migration. |
| Newman not found | Ensure `$(npm root -g)/../bin` is on PATH, or use `npx newman run ...` |

---

## 8. Related

- **Postman Desktop:** Import the same collection for manual testing — see [postman/README.md](../postman/README.md)
- **Full API reference:** [docs/API_REFERENCE.md](API_REFERENCE.md) (curl + responses)
