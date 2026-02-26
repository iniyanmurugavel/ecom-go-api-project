#!/usr/bin/env bash
# Run this from the project root. All credentials are read from .env.
# Usage: ./scripts/setup-and-run.sh

set -e
cd "$(dirname "$0")/.."

# Load all credentials from .env (single source of truth)
if [ -f .env ]; then
  set -a
  source .env
  set +a
  echo "==> Loaded .env"
else
  echo "==> No .env found; using defaults. Copy .env.example to .env to set credentials."
  export GOOSE_DRIVER=postgres
  export GOOSE_DBSTRING="host=localhost port=15432 user=postgres password=postgres dbname=ecom sslmode=disable"
  export GOOSE_MIGRATION_DIR=internal/adapters/postgresql/migrations
fi

echo "==> Starting PostgreSQL..."
docker compose up -d

echo "==> Waiting for Postgres to be ready..."
sleep 3

echo "==> Running migrations..."
goose up

echo "==> Seeding sample products..."
docker compose exec -T postgres psql -U postgres -d ecom -c "
  INSERT INTO products (name, price_in_centers, quantity)
  VALUES
    ('Sample Product A', 1999, 10),
    ('Sample Product B', 2999, 5),
    ('Sample Product C', 999, 20);
" 2>/dev/null || true

echo "==> Starting API server at http://localhost:8080"
echo "    Test in Postman: GET http://localhost:8080/health, GET /products, POST /orders"
echo "    Stop API: Ctrl+C. Stop DB: docker compose down (README: 'How to stop the API and the database')."
echo "    (If 8080 in use: HTTP_ADDR=:8081 go run cmd/*.go, then use :8081 in Postman)"
echo ""
go run cmd/*.go
