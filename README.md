## Ecommerce API (Go + PostgreSQL)

This is a small **e‑commerce API written in Go**.  
It is intentionally simple and is a great project for someone who has **just started learning Go** and wants to understand:

- **How to build an HTTP API in Go**
- **How to structure a small project** (folders, packages, `cmd/`, `internal/`, etc.)
- **How to talk to PostgreSQL from Go** using generated code (`sqlc`)
- **How to keep database schema in sync** using migrations (`goose`)

The API currently supports:

- **`GET /health`** – simple health check (body: "all good")
- **`GET /health/live`** – health + DB ping (returns 503 if DB is down)
- **`POST /v1/auth/register`** – create customer (email, password, name)
- **`POST /v1/auth/login`** – login and get JWT token
- **`GET /v1/products`** – list products (requires `Authorization: Bearer <token>`)
- **`POST /v1/orders`** – create order (requires JWT; customer ID from token, body has `items` only)

If you run into connection errors or “role postgres does not exist”, see **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** for a summary of common issues and how to fix them.

For **authentication (JWT)** — how to register, login, and use the token in Postman — see **[AUTH_README.md](AUTH_README.md)**.

For a **high-level design** of how the API works (request flow: API → handler → service → SQLC → PostgreSQL, Docker, and what to do when you make changes or want hot reload), see **[ARCHITECTURE.md](ARCHITECTURE.md)**.

For **development vs deployment** (use Docker or your own DB in dev? deploy in Docker later?), **Docker basics**, and a **command reference**, see **[DEVELOPMENT_AND_DEPLOYMENT.md](DEVELOPMENT_AND_DEPLOYMENT.md)**.

**If something is unclear:** The code has short comments (e.g. in `cmd/api.go` for middleware, `cmd/main.go` for startup). Use the links above for request flow, migrations, and running the project.

For a **summary of improvements** (connection pool, repository interfaces, tests, CI, etc.), see **[IMPROVEMENTS.md](IMPROVEMENTS.md)**.

**Scalability and testability (what the codebase does):** The API uses a **connection pool** (`pgxpool`) so many requests can use the DB at once; **repository interfaces** in `internal/products` and `internal/orders` so services can be tested with mock repos; **versioned routes** (`/v1/products`, `/v1/orders`) so you can add v2 later; **pagination** on list products (`?limit=&offset=`); **graceful shutdown** (Ctrl+C lets in-flight requests finish); **health with DB ping** (`/health/live`) for load balancers; and **rate limiting** per IP (configurable via `RATE_LIMIT_REQUESTS_PER_MINUTE`, default 100/min). Comments in the code explain each layer.

---

### Who this project is for

This project is for you if:

- You know **basic Go syntax** (variables, functions, structs, interfaces) but have never built a real API.
- You want to see a **practical, idiomatic** Go codebase that is still small enough to fully understand.
- You want to learn **how HTTP handlers, services, and database code fit together**.

You do **not** need to be an expert. You can learn by:

1. Running the code.
2. Calling the endpoints.
3. Reading the code with this README as a guide.

---

### Tech stack in one glance

- **Language**: Go (see `go.mod` for the exact version)
- **Web router**: [`chi`](https://github.com/go-chi/chi) – lightweight HTTP router
- **Database**: PostgreSQL
- **DB driver**: [`pgx`](https://github.com/jackc/pgx)
- **Migrations**: [`goose`](https://github.com/pressly/goose)
- **Query code generator**: [`sqlc`](https://docs.sqlc.dev/)
- **Container / local DB**: Docker + Docker Compose

---

### Project layout (high‑level map)

Here is the important folder/file map and **what you should read** as a beginner:

- `go.mod`  
  - Defines the **module name** and Go version.
  - Lists dependencies like `chi`, `pgx`, etc.

- `cmd/main.go`  
  - **Entry point** of the application (`main` function).
  - Reads configuration (DB DSN), sets up logging, connects to the database.
  - Creates the `application` struct and starts the HTTP server.

- `cmd/api.go`  
  - Defines the `application` type and config.
  - Sets up the **HTTP router** with `chi` and common middleware (logging, panic recovery, timeouts, etc.).
  - Wires **handlers** to paths:
    - `GET /health`
    - `GET /products`
    - `POST /orders`

- `internal/env/env.go`  
  - Tiny helper to read **environment variables with a fallback**.

- `internal/json/json.go`  
  - Helpers to **read JSON from requests** and **write JSON responses**.

- `internal/products/`  
  - `service.go`: business logic for products (currently just listing products).
  - `handlers.go`: HTTP handler that calls the service and returns JSON.

- `internal/orders/`  
  - `types.go`: request types (`createOrderParams`, `orderItem`) and the `Service` interface.
  - `service.go`: business logic for placing an order.
    - Starts a DB transaction.
    - Validates input.
    - Creates an order.
    - For each item, checks product stock and creates order items.
  - `handlers.go`: HTTP handler that reads JSON, calls the service, and returns JSON or errors.

- `internal/adapters/postgresql/`  
  - `migrations/` – **SQL migration files** that create the `products`, `orders`, and `order_items` tables.
  - `sqlc/` – **generated Go code** that wraps SQL queries:
    - `models.go` – Go structs like `Product`, `Order`, `OrderItem`.
    - `queries.sql.go` – Go functions generated from `.sql` queries (e.g. `ListProducts`, `FindProductByID`, etc.).
    - `db.go`, `querier.go` – helpers and interfaces so the rest of the code can call repo methods.

As a learner, a good reading order is:

1. `cmd/main.go`
2. `cmd/api.go`
3. `internal/products/service.go` and `internal/products/handlers.go`
4. `internal/orders/types.go`, `internal/orders/service.go`, `internal/orders/handlers.go`
5. `internal/adapters/postgresql/sqlc/*`

---

### How the API works (high‑level)

In simple words:

1. **`main` connects to PostgreSQL** and sets up logging.
2. It builds an `application` value with:
   - configuration (`addr`, DB connection string),
   - a live DB connection.
3. It calls `app.mount()` to build the **HTTP router**:
   - configures middleware (logging, recovery, timeouts),
   - wires handlers to endpoints.
4. It starts an `http.Server` that listens on `:8080`.
5. When a request comes in:
   - `chi` finds the matching route.
   - The route calls a **handler** (e.g. `ListProducts`, `PlaceOrder`).
   - The handler calls a **service** (business logic).
   - The service uses the **repository (`sqlc`)** to talk to PostgreSQL.
   - The handler writes the response as JSON.

---

### Prerequisites

You should have:

- **Go** installed (version compatible with the one in `go.mod`).
- **Docker** and **Docker Compose**.
- **Goose** (DB migrations).
- **SQLC** (code generator).

---

### Setup necessary tooling (step‑by‑step)

Follow these steps once to install and verify everything you need.

| Step | What to do | Understand & do for future learning |
|------|------------|--------------------------------------|
| **1. Install Go** | Download and install from [go.dev](https://go.dev/dl/). Check: `go version` | Go is the language and runtime. You’ll use `go run`, `go build`, and `go install` often. |
| **2. Install Docker** | Install [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or Docker Engine on Linux). Check: `docker --version` and `docker compose version` | Docker runs PostgreSQL in a container so you don’t install it on your machine. `docker compose` reads `docker-compose.yaml`. |
| **3. Install Goose** | Run: `go install github.com/pressly/goose/v3/cmd/goose@latest` | Goose applies SQL migrations (e.g. creating tables). The binary goes to `$(go env GOPATH)/bin`. |
| **4. Install SQLC** | Run: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` | SQLC generates type-safe Go code from your SQL. You run `sqlc generate` after changing queries. |
| **5. Add Go bin to PATH** | Ensure `$(go env GOPATH)/bin` is on your PATH. Check: `go env GOPATH` then run `goose -h` and `sqlc version` | If `goose` or `sqlc` is not found, add that bin directory to your shell’s PATH (e.g. in `~/.zshrc` or `~/.bashrc`). |

**Commands to run (copy-paste):**

```bash
# 1. Verify Go
go version

# 2. Verify Docker
docker --version
docker compose version

# 3. Install Goose (migrations)
go install github.com/pressly/goose/v3/cmd/goose@latest

# 4. Install SQLC (query code generator)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# 5. Ensure Go’s bin is on PATH (Unix/macOS). Add to ~/.zshrc or ~/.bashrc if needed:
export PATH="$PATH:$(go env GOPATH)/bin"

# 6. Verify tools
goose -h
sqlc version
```

After this, you should be able to run `goose`, `sqlc`, `docker`, and `go` from the terminal.

---

### Using .env for all credentials

All credentials and config (database URL, HTTP port) are read from a single **`.env`** file so you don’t hardcode secrets or repeat them in scripts.

1. **Create your `.env`** (once per machine):
   ```bash
   cp .env.example .env
   ```
   Edit `.env` if you need different values (e.g. DB password, port).

2. **What uses `.env`**
   - The **Go app** loads `.env` at startup (`godotenv`) and uses `GOOSE_DBSTRING` (or `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`) and `HTTP_ADDR`.
   - The **setup script** (`./scripts/setup-and-run.sh`) sources `.env` before running migrations and the app, so Goose and the API use the same credentials.

3. **Do not commit `.env`** (it’s in `.gitignore`). Commit `.env.example` as a template; everyone copies it to `.env` and fills in their own values.

**One database for the project** — The app uses a single database: Postgres running in Docker. There is only one DB; the API, migrations (Goose), and any DB client (e.g. TablePlus) all connect to it. If you have another Postgres on your machine (e.g. on port 5432), that is separate; this project uses only the Docker one on port 15432.

**Standard Docker Postgres connection** (app, Goose, and any DB client use these):

| Host     | Port   | Database | User     | Password | SSL  |
|----------|--------|----------|----------|----------|------|
| localhost | 15432 | ecom     | postgres | postgres | off  |

---

### Run the project (step‑by‑step)

| Step | What to do | Understand & do for future learning |
|------|------------|--------------------------------------|
| **1. Clone** | `git clone <repo-url>` then `cd ecom-go-api-project` | You need the code on your machine. `cd` into the project root for all following commands. |
| **2. Create .env** | `cp .env.example .env` (edit if needed) | All credentials live in `.env`; the app and scripts read from it. |
| **3. Start DB** | `docker compose up -d` | Starts PostgreSQL in a container. The app connects using the DSN from `.env` (default port 15432). |
| **4. Migrate** | Either run `./scripts/setup-and-run.sh` (sources `.env` and runs migrations + app) or `source .env; goose up` | Creates tables from `migrations/`. Script uses `.env` so no need to export vars by hand. |
| **5. Generate SQLC (if needed)** | `sqlc generate` | Only needed after you change `.sql` queries or schema. Generates Go code in `internal/adapters/postgresql/sqlc/`. |
| **6. Start API** | `go run cmd/*.go` | Runs the app (reads `.env`); server listens on `http://localhost:8080`. Stop with Ctrl+C. |

**Commands to run (in order):**

```bash
# 1. Clone and enter project
git clone <this-repo-url>
cd ecom-go-api-project

# 2. Create .env from template (all credentials go here)
cp .env.example .env
# Edit .env if you need different DB host/port/user/password

# 3. Start PostgreSQL
docker compose up -d

# 4. Run migrations and start API (script sources .env automatically)
./scripts/setup-and-run.sh
# Or manually: source .env; goose up; go run cmd/*.go

# 5. Regenerate SQLC only if you changed queries/schema
sqlc generate
```

You should see:

```text
server has started at addr :8080
```

The API is now available at **`http://localhost:8080`**. Keep this terminal open while you test.

---

### How to stop the API and the database

| What to stop | Command | Comment |
|--------------|---------|---------|
| **API** | Press **Ctrl+C** in the terminal where `go run cmd/*.go` is running | Stops the Go server only. The database keeps running. |
| **Database (Docker)** | `docker compose down` | Stops and removes the Postgres container. Your data stays in the Docker volume (safe for next `docker compose up -d`). |
| **Database and delete all data** | `docker compose down -v` | Stops the container and **removes the volume** — next time you start, you get an empty DB (run migrations again). |
| **Check if DB is running** | `docker ps` | You should see `ecom-postgres` when the DB is up. |

So: use **Ctrl+C** to stop the API; use **`docker compose down`** when you want to stop the DB (e.g. at end of day). Start again with `docker compose up -d` then `go run cmd/*.go`.

---

### Testing with Postman (step‑by‑step)

Use [Postman](https://www.postman.com/downloads/) to send requests and inspect responses. Base URL for all requests: **`http://localhost:8080`**.

| Step | What to do in Postman | Understand & do for future learning |
|------|------------------------|--------------------------------------|
| **1. Open Postman** | Install Postman (if needed), open it, and create a new request (e.g. New → HTTP Request). | Postman lets you set method, URL, headers, and body without writing code. Same ideas apply to any HTTP client (curl, browser, frontend). |
| **2. Test health** | Set **Method** to `GET`, **URL** to `http://localhost:8080/health`. Click **Send**. | You should get status `200` and body `all good`. Use `GET /health/live` to also ping the DB (503 if DB is down). |
| **3. Register & login** | See **[AUTH_README.md](AUTH_README.md)** for full steps. Summary: `POST /v1/auth/register` with `{ email, password, name }`, then `POST /v1/auth/login` to get a token. | Copy the `token` from the login response. |
| **4. Test list products** | Set **Method** to `GET`, **URL** to `http://localhost:8080/v1/products`. In **Headers** add `Authorization: Bearer <your-token>`. Click **Send**. | You get JSON: an array of products. Without token you get `401`. |
| **5. Test place order** | Set **Method** to `POST`, **URL** to `http://localhost:8080/v1/orders`. **Headers:** `Authorization: Bearer <token>`, `Content-Type: application/json`. **Body (raw, JSON):** `{ "items": [{ "productId": 1, "quantity": 2 }] }`. Click **Send**. | You get `201 Created`. Customer ID comes from the token (not from body). |

**Postman setup per request:**

| Endpoint | Method | URL | Headers | Body |
|----------|--------|-----|---------|------|
| Health check | `GET` | `http://localhost:8080/health` | (none) | (none) |
| Health (with DB) | `GET` | `http://localhost:8080/health/live` | (none) | (none) |
| Register | `POST` | `http://localhost:8080/v1/auth/register` | `Content-Type: application/json` | `{ "email", "password", "name" }` |
| Login | `POST` | `http://localhost:8080/v1/auth/login` | `Content-Type: application/json` | `{ "email", "password" }` |
| List products | `GET` | `http://localhost:8080/v1/products` | `Authorization: Bearer <token>` | Optional: `?limit=20&offset=0` |
| Place order | `POST` | `http://localhost:8080/v1/orders` | `Authorization: Bearer <token>`, `Content-Type: application/json` | `{ "items": [{ "productId", "quantity" }] }` |

**Body for Place order (raw JSON):** Customer ID comes from the JWT — do **not** send `customerId` in the body.

```json
{
  "items": [
    { "productId": 1, "quantity": 2 },
    { "productId": 2, "quantity": 1 }
  ]
}
```

Use real `productId` values from the **List products** response. Get a token first via **Login** (see [AUTH_README.md](AUTH_README.md)).

**What to check in Postman:**

- **Status**: `200` (health, products), `201` (order created), `400` (bad request), `404` (product not found), `500` (server error).
- **Response body**: JSON for `/products` and `/orders`; plain text for `/health`.
- **Headers**: Response may include `Content-Type: application/json` for JSON endpoints.

**For future learning:** Try changing `productId` to a non-existent ID and see `404`; send invalid JSON and see `400`; then find in the code where these status codes are set (`internal/orders/handlers.go`, `internal/json`).

---

### HTTP endpoints and examples

Once the server is running, you can use **Postman** (see [Testing with Postman](#testing-with-postman-step‑by‑step) above), `curl`, or any HTTP client.

#### `GET /health`

Quick check that the server is alive.

```bash
curl http://localhost:8080/health
# With DB check: curl http://localhost:8080/health/live
```

Response (plain text):

```text
all good
```

#### `POST /v1/auth/register` and `POST /v1/auth/login`

Register creates a customer; login returns a JWT. See **[AUTH_README.md](AUTH_README.md)** for full details.

```bash
# Register
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"secret123","name":"Your Name"}'

# Login (save the token from response)
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"secret123"}'
```

#### `GET /products`

Returns all products from the `products` table as JSON. **Requires JWT.**

```bash
# First login to get TOKEN, then:
curl "http://localhost:8080/v1/products?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```

Example response:

```json
[
  {
    "id": 1,
    "name": "Example product",
    "price_in_centers": 1000,
    "quantity": 5,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

This handler lives in `internal/products/handlers.go` and calls the service in `internal/products/service.go`.

#### `POST /orders`

Creates a new order. **Customer ID comes from the JWT** (not from the body). Body has only `items`.

Request:

```bash
# TOKEN from login response
curl -X POST http://localhost:8080/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      { "productId": 1, "quantity": 2 },
      { "productId": 2, "quantity": 1 }
    ]
  }'
```

If all products exist and have enough stock, you get back the created order:

```json
{
  "id": 10,
  "customer_id": 1,
  "created_at": "2024-01-01T00:00:00Z"
}
```

If a product ID does not exist, the service returns `ErrProductNotFound` and the handler sends:

- HTTP status: `404 Not Found`

If something else goes wrong, you get:

- HTTP status: `500 Internal Server Error`

The request/response types and logic live in:

- `internal/orders/types.go`
- `internal/orders/service.go`
- `internal/orders/handlers.go`

---

### Deep dive: request flow in code (for learners)

Let’s walk through what happens for `GET /products`:

1. **`main.go`**:
   - Creates `cfg` (address + DB config).
   - Connects to Postgres using `pgx.Connect`.
   - Creates an `application` with `config` and `db`.
   - Calls `api.run(api.mount())`.

2. **`api.go` – `mount()`**:
   - Builds a `chi.NewRouter()`.
   - Adds middlewares:
     - `RequestID`, `RealIP`, `Logger`, `Recoverer`, `Timeout`.
   - Creates the product service:
     - `productService := products.NewService(repo.New(app.db))`
   - Creates the handler:
     - `productHandler := products.NewHandler(productService)`
   - Registers the route:
     - `r.Get("/v1/products", productHandler.ListProducts)`

3. **`products/handlers.go` – `ListProducts` handler**:
   - Calls the service:
     - `products, err := h.service.ListProducts(r.Context())`
   - Handles errors (500 on failure).
   - Writes JSON to the client:
     - `json.Write(w, http.StatusOK, products)`

4. **`products/service.go` – `ListProducts` service**:
   - Has a field `repo repo.Querier` which is implemented by the generated `repo.Queries`.
   - Calls:
     - `return s.repo.ListProducts(ctx)`

5. **`sqlc/queries.sql.go` – `ListProducts`**:
   - Runs the actual SQL:
     - `SELECT id, name, price_in_centers, quantity, created_at FROM products`
   - Scans each row into a `Product` struct.
   - Returns `[]Product` to the service.

6. **Back up the stack**:
   - Service returns the slice to the handler.
   - Handler writes it as JSON.
   - The client gets the JSON response.

For `POST /orders` the flow is similar, but with an additional **transaction**:

- `orders/handlers.go`:
  - Reads JSON body into `createOrderParams` using `json.Read`.
  - Calls `service.PlaceOrder`.
- `orders/service.go`:
  - Validates input.
  - Starts a transaction: `tx, err := s.db.Begin(ctx)`.
  - Calls `qtx := s.repo.WithTx(tx)` so that all queries use the same transaction.
  - Creates an order and order items, checks product stock.
  - On success: `tx.Commit(ctx)`.
  - On failure: `defer tx.Rollback(ctx)` undoes the partial work.

This is a common pattern in Go:

- Handlers handle **HTTP concerns** (status codes, headers, JSON).
- Services handle **business rules** (validation, transactions).
- Repositories (here: generated `sqlc` code) handle **SQL**.

---

### Deep dive: database + `sqlc`

- SQL migrations in `internal/adapters/postgresql/migrations` create tables like:
  - `products`
  - `orders`
  - `order_items`
- `sqlc` reads your `.sql` files and generates:
  - Go structs that map to rows (`Product`, `Order`, `OrderItem`).
  - Go methods for each query, e.g.:
    - `CreateOrder`
    - `CreateOrderItem`
    - `FindProductByID`
    - `ListProducts`

Because of this, the service code looks like **regular Go** and does not have to manually scan database rows everywhere.

If you add a new feature that needs a new query:

1. Add the SQL to the appropriate `.sql` file (following the `-- name: ...` convention).
2. Run `sqlc generate`.
3. Use the new generated method from your service.

---

### Adding a new table (step‑by‑step)

Let’s say you want to add a new table (for example, `customers`):

1. **Create a new migration file** under `internal/adapters/postgresql/migrations`:

   ```bash
   goose -dir internal/adapters/postgresql/migrations create create_customers sql
   ```

   Edit the generated file and write the SQL to create your table.

2. **Run the migrations**:

   ```bash
   goose -dir internal/adapters/postgresql/migrations postgres "$GOOSE_DBSTRING" up
   ```

3. **Add queries for `sqlc`** (if needed) and regenerate:

   ```bash
   sqlc generate
   ```

4. **Consume it from your services** 🚀  
   - Add new methods to your service interface.
   - Call the generated `repo` methods from your service.
   - Expose new HTTP endpoints via handlers.

---

### If you are new to Go – how to study this project

Suggested learning path:

- **Step 1** – Run the project and hit `/health`, `/products`, `/orders`.
- **Step 2** – Read `cmd/main.go` and answer:
  - How is the DB connection string built?
  - Where does the HTTP server start?
- **Step 3** – Read `cmd/api.go` and answer:
  - How are routes registered?
  - What middlewares are used?
- **Step 4** – Read `internal/products/service.go` and `internal/products/handlers.go`:
  - How does a handler call a service?
  - How does the service call the repo?
- **Step 5** – Read `internal/orders/service.go`:
  - Why do we use a transaction here?
  - What happens when stock is insufficient?
- **Step 6** – Open the migration files and generated `sqlc` code to see how SQL maps to Go.

If you go through these steps slowly and experiment (change log messages, add fields, break things and fix them), you will build a very strong foundation for writing Go backends.
