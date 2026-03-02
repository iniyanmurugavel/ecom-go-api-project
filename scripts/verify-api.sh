#!/usr/bin/env bash
# Verify all API endpoints work. Run with API already running (e.g. ./scripts/setup-and-run.sh).
# Usage: ./scripts/verify-api.sh [BASE_URL]
# Exit 0 if all pass, 1 otherwise.

set -e
BASE="${1:-http://localhost:8080}"
EMAIL="verify$(date +%s)@example.com"

echo "==> Verifying API at $BASE"
echo ""

# 1. Health
echo "1. GET /health"
if ! curl -sf "$BASE/health" | grep -q '"status":"ok"'; then
  echo "   FAIL: expected JSON with status ok"
  exit 1
fi
echo "   OK"

# 2. Health/live
echo "2. GET /health/live"
if ! curl -sf "$BASE/health/live" | grep -q "ok"; then
  echo "   FAIL: expected 'ok'"
  exit 1
fi
echo "   OK"

# 3. Metrics
echo "3. GET /metrics"
if ! curl -sf "$BASE/metrics" | grep -q "go_"; then
  echo "   FAIL: expected Prometheus metrics"
  exit 1
fi
echo "   OK"

# 4. Register
echo "4. POST /v1/auth/register"
REG=$(curl -sf -X POST "$BASE/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"secret123\",\"name\":\"Test User\"}")
if ! echo "$REG" | grep -q '"id"'; then
  echo "   FAIL: $REG"
  exit 1
fi
echo "   OK"

# 5. Login
echo "5. POST /v1/auth/login"
LOGIN=$(curl -sf -X POST "$BASE/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"secret123\"}")
TOKEN=$(echo "$LOGIN" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
  echo "   FAIL: no token in $LOGIN"
  exit 1
fi
echo "   OK"

# 6. Products (auth required)
echo "6. GET /v1/products"
PRODS=$(curl -sf "$BASE/v1/products?limit=2" -H "Authorization: Bearer $TOKEN")
if ! echo "$PRODS" | grep -q '"id"'; then
  echo "   FAIL: $PRODS"
  exit 1
fi
echo "   OK"

# 7. Place order
echo "7. POST /v1/orders"
ORDER=$(curl -sf -X POST "$BASE/v1/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"productId":1,"quantity":1}]}')
if ! echo "$ORDER" | grep -q '"id"'; then
  echo "   FAIL: $ORDER"
  exit 1
fi
echo "   OK"

echo ""
echo "==> All 7 endpoints verified successfully"
