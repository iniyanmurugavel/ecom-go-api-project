# JWT Security Improvements

This document explains the **current weaknesses** in our JWT setup and **how to improve** them, from quick wins to advanced patterns.

---

## 1. Current Weaknesses

| Issue | Risk | Current state |
|-------|------|----------------|
| **Weak secret** | Attacker can forge tokens if secret is guessable | We warn if < 32 chars but app still runs |
| **No secret enforcement** | Dev might deploy with default secret | No fail-fast on weak secret |
| **Long-lived token (24h)** | Stolen token valid for 24 hours | Single token, no refresh |
| **No issuer/audience** | Token could be reused across services | Only customer_id, email, exp, iat |
| **No token revocation** | Can't invalidate on logout/password change | Stateless = no server-side revoke |
| **HS256 only** | Shared secret; all services need it | Fine for single API; RS256 better for microservices |

---

## 2. Quick Wins (Easy to Implement) ✅ IMPLEMENTED

### 2.1 Enforce Strong Secret in Production ✅

**Problem:** App starts with weak `JWT_SECRET`.

**Fix:** When `APP_ENV=production`, app exits if `JWT_SECRET` < 32 chars. Set `APP_ENV=development` for local dev.

**Env:** `APP_ENV`, `JWT_SECRET`

---

### 2.2 Add Issuer and Audience Claims ✅

**Problem:** Token could be misused if copied to another service.

**Fix:** Added `iss` and `aud` claims. Verified on parse via `jwt.WithIssuer` and `jwt.WithAudience`.

**Env:** `JWT_ISSUER` (default: ecom-api), `JWT_AUDIENCE` (default: ecom-api)

---

### 2.3 Configurable Token Expiry ✅

**Problem:** 24h was hardcoded.

**Fix:** Read from `JWT_EXPIRY_HOURS` (default: 24).

---

## 3. Medium Effort: Refresh Tokens

**Problem:** Long-lived access token = big window if stolen.

**Fix:** Short-lived access token (e.g. 15 min) + long-lived refresh token (e.g. 7 days).

**Flow:**
1. Login returns `access_token` (15 min) + `refresh_token` (7 days, stored in DB or signed with different secret).
2. Client uses `access_token` for API calls.
3. When access token expires, client sends `refresh_token` to `POST /v1/auth/refresh`.
4. Server validates refresh token, issues new access token (and optionally new refresh token).

**Benefits:** Stolen access token expires in 15 min. Refresh token can be revoked (if stored in DB).

**Implementation:** Add `refresh_tokens` table (token hash, user_id, expires_at, revoked). Or sign refresh token with a different secret and store hash in DB for revocation.

---

## 4. Token Revocation (Logout / Password Change)

**Problem:** Stateless JWT can't be "cancelled" before expiry.

**Options:**

| Approach | How | Pros | Cons |
|----------|-----|------|------|
| **Short expiry + refresh** | Access token 15 min; revoke refresh token in DB | Revocation works | Need refresh flow |
| **Blocklist (blacklist)** | Store revoked token IDs (jti) in Redis/DB until expiry | Works with current design | Need Redis/DB; check on every request |
| **Version in token** | Add `token_version` in JWT; increment on password change; check version in DB | Simple | DB lookup per request (or cache) |

**Recommendation:** For this project, short expiry + refresh tokens + revoke refresh token on logout/password change is the cleanest.

---

## 5. RS256 (Asymmetric) for Microservices

**Problem:** With HS256, every service that verifies tokens needs the secret. Leak = game over.

**Fix:** Use RS256 (RSA). One service has the **private key** (signs tokens). Others have **public key** (verify only).

**Flow:**
- Auth service: holds private key, signs tokens on login.
- API services: hold public key, verify tokens (can't forge).

**When to use:** Multiple services (auth service + API services). Overkill for a single API.

---

## 6. Recommended Improvement Order

| Priority | Improvement | Effort | Impact |
|----------|-------------|--------|--------|
| 1 | Enforce JWT_SECRET length in production | Low | High |
| 2 | Add iss/aud claims | Low | Medium |
| 3 | Configurable expiry (env var) | Low | Low |
| 4 | Refresh tokens + short access token | Medium | High |
| 5 | Revoke refresh token on logout | Medium | High |
| 6 | RS256 (if multi-service) | Medium | Medium |

---

## 7. Generate a Strong Secret

```bash
# 32 bytes = 256 bits, good for HS256
openssl rand -base64 32
```

Put the output in `JWT_SECRET` in production. Never commit it.

---

## 8. Summary

**Minimum for production:**
- JWT_SECRET ≥ 32 chars, fail if not
- Add iss/aud
- Use `openssl rand -base64 32` for secret

**Better:**
- Refresh tokens + 15 min access token
- Revoke refresh token on logout

**Advanced:**
- RS256 if you have multiple services
