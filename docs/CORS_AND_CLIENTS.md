# CORS and API Clients: Web vs Mobile Explained

This document explains **what CORS is**, **why browsers need it**, **why mobile apps don't**, and how to configure this API for different clients. Includes concrete examples.

---

## 1. What Is CORS?

**CORS** = Cross-Origin Resource Sharing.

It is a **browser security rule**, not an HTTP or API rule. Only web browsers enforce it.

**The rule:** When a web page at `http://localhost:3000` (origin A) makes a request to `http://localhost:8080` (origin B), the browser blocks the response **unless** the API at origin B explicitly says "I allow requests from origin A."

**Origin** = protocol + domain + port. So:
- `http://localhost:3000` and `http://localhost:8080` are **different origins**.
- `https://api.example.com` and `https://app.example.com` are **different origins**.

---

## 2. Why Does the Browser Do This?

Without CORS, any website could secretly call your bank API from the user's browser (using the user's cookies) and steal data. CORS forces the **API** to explicitly allow which origins can call it.

**Example of the problem CORS solves:**

```
BAD (no CORS):
  - User visits evil-site.com
  - evil-site.com runs JavaScript: fetch("https://your-bank.com/transfer?to=hacker")
  - Browser sends request WITH user's bank cookies
  - Bank transfers money
  - evil-site.com reads the response

GOOD (with CORS):
  - Bank API responds with: "Access-Control-Allow-Origin: evil-site.com" → NO (bank doesn't allow evil-site)
  - Browser blocks the response from reaching evil-site's JavaScript
  - User's data stays safe
```

---

## 3. Example: React Frontend Calling This API

### Setup

- **React app** runs at `http://localhost:3000`
- **This API** runs at `http://localhost:8080`

### Without CORS configured

```javascript
// In your React app (http://localhost:3000)
fetch('http://localhost:8080/v1/products', {
  headers: { 'Authorization': 'Bearer ' + token }
})
.then(res => res.json())
.then(data => console.log(data));
```

**What happens:**

1. Browser sends the request to `http://localhost:8080`
2. API responds with 200 and JSON
3. **Browser blocks** the response from reaching your JavaScript
4. You see in DevTools: `Access to fetch at 'http://localhost:8080/v1/products' from origin 'http://localhost:3000' has been blocked by CORS policy`

**Why?** The API did not send `Access-Control-Allow-Origin: http://localhost:3000` in the response. Browser says: "Different origin, no permission → block."

### With CORS configured

**1. Add to your `.env`:**

```
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

**2. Restart the API** (`go run ./cmd`)

**3. Same React code now works:**

```javascript
fetch('http://localhost:8080/v1/products', {
  headers: { 'Authorization': 'Bearer ' + token }
})
.then(res => res.json())
.then(data => console.log(data));  // ✅ Works! You get the data.
```

**What happens:**

1. Browser sends request (same as before)
2. API responds with 200, JSON, **and** header: `Access-Control-Allow-Origin: http://localhost:3000`
3. Browser sees: "API allows this origin" → **lets your JavaScript read the response**
4. Your React app gets the data

---

## 4. Example: Mobile App (iOS/Android) Calling This API

### Setup

- **Mobile app** (e.g. React Native, Flutter, Swift, Kotlin) makes HTTP request
- **This API** runs at `http://10.0.2.2:8080` (Android emulator) or `http://localhost:8080` (iOS simulator) or `https://api.yourdomain.com` (production)

### Mobile code (conceptual)

```javascript
// React Native example
fetch('http://localhost:8080/v1/products', {
  headers: { 'Authorization': 'Bearer ' + token }
})
.then(res => res.json())
.then(data => console.log(data));
```

**What happens:**

1. The **mobile app** (not a browser) sends the request
2. API responds with 200 and JSON
3. **No CORS check** — mobile runtimes (React Native, Flutter, native iOS/Android) do **not** enforce CORS
4. Your app gets the data ✅

**Why no CORS?** CORS is enforced by **web browsers** to protect users from malicious websites. Mobile apps are installed apps; they don't have the same "any website can run code" model. The OS trusts the app.

---

## 5. Summary Table

| Client type              | Runs in        | CORS enforced? | Need CORS_ALLOWED_ORIGINS? |
|--------------------------|----------------|----------------|----------------------------|
| React/Vue/Angular (web)  | Browser        | Yes            | Yes — add your frontend URL |
| Postman                  | Desktop app    | No             | No                          |
| curl                     | Terminal       | No             | No                          |
| Mobile app (React Native)| App runtime    | No             | No                          |
| Mobile app (Flutter)     | App runtime    | No             | No                          |
| Native iOS (Swift)       | App            | No             | No                          |
| Native Android (Kotlin)  | App            | No             | No                          |

---

## 6. When to Set CORS_ALLOWED_ORIGINS

| Scenario                                      | What to do |
|-----------------------------------------------|------------|
| Only testing with Postman/curl                | Leave empty (CORS disabled) |
| Building a React frontend at localhost:3000    | `CORS_ALLOWED_ORIGINS=http://localhost:3000` |
| Production: frontend at https://app.example.com| `CORS_ALLOWED_ORIGINS=https://app.example.com` |
| Multiple frontends                            | `CORS_ALLOWED_ORIGINS=https://app1.com,https://app2.com` |
| Mobile app only                               | Leave empty (not needed) |

---

## 7. Where CORS Is Configured in This Project

- **File:** `cmd/api.go`
- **Package:** `github.com/go-chi/cors`
- **Config:** `app.config.corsOrigins` from `CORS_ALLOWED_ORIGINS` env var (parsed in `cmd/main.go`)

When `CORS_ALLOWED_ORIGINS` is set, the CORS middleware adds `Access-Control-Allow-Origin` to responses for allowed origins.

---

## 8. Quick Test

**Test without CORS (Postman):** Always works — Postman is not a browser.

**Test with CORS (browser):** Open DevTools Console on `http://localhost:3000` and run:

```javascript
fetch('http://localhost:8080/health').then(r => r.text()).then(console.log);
```

- Without CORS: You may see a CORS error (if your page origin ≠ API origin).
- With `CORS_ALLOWED_ORIGINS=http://localhost:3000`: You see `"all good"`.
