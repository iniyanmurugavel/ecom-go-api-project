# Development and Deployment: Basics for Beginners

This file answers: **Should I use Docker or my own DB in development?** and **Do I deploy in Docker later?** It also explains Docker basics and gives you a simple command reference. If you're new to Docker or deployment, read this once and keep it as reference.

---

## 1. Development on your machine: Docker for DB, API on your machine (recommended)

### What this project does today

| What | Where it runs in development |
|------|------------------------------|
| **PostgreSQL (database)** | Inside **Docker** (container `ecom-postgres`). You start it with `docker compose up -d`. |
| **Your Go API** | On **your machine** (not in Docker). You run `go run ./cmd` in the terminal. |

So in dev:

- **Database** → Docker (so you don’t install Postgres on your Mac and avoid port clashes with other tools).
- **API** → Your machine (easy to edit code, run tests, use debugger).

You **do not** need to install PostgreSQL on your Mac for this project. Docker gives you a Postgres that’s isolated and matches the same setup you can use later (e.g. in production or in a full Docker deploy).

### Can I use my own PostgreSQL (installed on my Mac) instead of Docker?

Yes. If you already have Postgres (e.g. Postgres.app, Homebrew) and want to use it:

1. Create a database named `ecom` and a user `postgres` with password `postgres` (or whatever you prefer).
2. In `.env`, set `GOOSE_DBSTRING` (and optionally `DB_*`) to point to **your** Postgres Use port 15432. To make Mac Postgres use 15432, see **[docs/MAC_POSTGRES_15432.md](docs/MAC_POSTGRES_15432.md)**.
3. Run migrations: `goose up`.
4. Run the API: `go run ./cmd`.

The app doesn’t care whether the DB is in Docker or on your host; it only cares about the connection string in `.env`. Using Docker is recommended so everyone has the same setup and you avoid “works on my machine” issues.


### How to reflect changes in your local DB

"Reflect in my local DB" means: make the database (tables, columns, data) match what the app expects. You do that by running **migrations** (and optionally seeding). Which DB is used depends on `.env`.

**If your local DB is Docker (default):**

1. Start DB: `docker compose up -d`
2. In `.env`: `GOOSE_DBSTRING="host=localhost port=15432 user=postgres password=postgres dbname=ecom sslmode=disable"`
3. Apply migrations: `source .env` then `goose up`
4. Optional seed: `docker exec ecom-postgres psql -U postgres -d ecom -c "INSERT INTO products (name, price_in_centers, quantity) VALUES ('Sample Product A', 1999, 10), ('Sample Product B', 2999, 5), ('Sample Product C', 999, 20);"`
5. See data: use a DB client with **localhost, port 15432**, database **ecom** — or run `docker exec ecom-postgres psql -U postgres -d ecom -c "SELECT * FROM products;"`

**If your local DB is Postgres on your Mac (e.g. Postgres.app):**

1. Configure Mac Postgres to use port **15432** (see "Make Mac Postgres use port 15432" above).
2. Create database `ecom` and user `postgres` with password `postgres`.
3. In `.env`: `GOOSE_DBSTRING="host=localhost port=15432 user=postgres password=postgres dbname=ecom sslmode=disable"`
4. Apply migrations: `source .env` then `goose up`
5. Run API: `go run ./cmd` — app uses the same `.env`, so it talks to your Mac Postgres.
6. See data: DB client → **localhost, port 15432**, database **ecom**.

**Summary:** Use port **15432** for both Docker and Mac Postgres. Then run `goose up`. That DB is your local DB; schema and data reflect there.

### TablePlus (or any DB client) and Docker: same DB

**Yes – both your API and TablePlus can use the same database.** There is only one Postgres; Docker runs it. Your API and TablePlus are just two different **clients** connecting to that same server.

- **Docker** = runs the database (PostgreSQL).
- **Your API** = connects to it using `.env` (host localhost, port 15432, database ecom).
- **TablePlus** = connect to the **same** database with the same settings.

**In TablePlus (or any DB client):** Use the same connection as the app (see README: “Standard Docker Postgres connection”):

| Host     | Port   | Database | User     | Password | SSL  |
|----------|--------|----------|----------|----------|------|
| localhost | 15432 | ecom     | postgres | postgres | off  |

Save and connect. You will see the same `products`, `orders`, and `order_items` that the API uses. Any change in TablePlus or via the API is in this one database; both stay in sync.

---

## 2. Later: deploying “in Docker”

“Deploy in Docker” usually means running your **API** and/or **database** inside containers on a server (or in the cloud).

### Two common cases

**A) Only the database in Docker (what you do in dev now)**  
- DB runs in a container.  
- API runs on the host (your laptop or a server).  
- Good for: local dev, or a server where the API is installed and only the DB is containerized.

**B) Both API and DB in Docker**  
- DB in one container (e.g. Postgres).  
- API in another container (your Go app).  
- They talk over a Docker network (e.g. API connects to `postgres:5432` inside the network, not `localhost:15432`).  
- Good for: production, CI/CD, or “run everything with one `docker compose up`.”

Right now this repo only defines a **Postgres** container. To “deploy the API in Docker” you would add a second service in `docker-compose.yaml` that builds and runs your Go binary. That’s a next step; for learning, sticking with **DB in Docker, API on your machine** is enough.

### Summary

- **Development:** Use **Docker for the DB** (recommended); run the **API on your machine** with `go run ./cmd`.  
- **Later / production:** You can deploy **only the DB** in Docker, or **both API and DB** in Docker; the app just needs the right connection string (host/port) for where Postgres is running.

---

## 3. Docker basics (minimal you need to know)

### What is Docker?

- **Image:** A snapshot of a system (e.g. “PostgreSQL 16”). It doesn’t run by itself.  
- **Container:** A **running instance** of an image. You can have many containers from the same image.  
- **Docker Compose:** A way to define and run multiple containers (e.g. one for Postgres) using a file `docker-compose.yaml`.

For this project you only run **one** container: Postgres. The file `docker-compose.yaml` describes it (image, port, password, volume).

### Why use Docker for the database?

- You don’t install Postgres on your Mac.  
- Same setup for everyone (same image, same port).  
- Easy to reset: `docker compose down -v` and `docker compose up -d` gives a fresh DB.  
- Later you can run the same image in production or in CI.

---

## 4. Commands you need (development)

Run these from the **project root** (`ecom-go-api-project`).

### Docker (database)

| Command | What it does |
|--------|----------------|
| `docker compose up -d` | Start Postgres in the background. Use this **first** so the DB is running. |
| `docker compose down` | Stop and remove the Postgres container. Data in the volume is kept. |
| `docker compose down -v` | Stop the container **and delete the data volume** (fresh DB next time you `up`). |
| `docker ps` | List running containers. You should see `ecom-postgres` when the DB is up. |
| `docker compose logs -f postgres` | Show Postgres logs (Ctrl+C to stop). |
| `docker exec ecom-postgres psql -U postgres -d ecom -c "SELECT 1;"` | Run a SQL command inside the container (quick test that DB is up). |

### Your app (API)

| Command | What it does |
|--------|----------------|
| `go run ./cmd` | Build and run the API. It reads `.env` and connects to Postgres (port 15432). Stop with Ctrl+C. |
| `go test -v ./...` | Run all tests. Integration tests skip if DB is unavailable. |
| `go test -v ./cmd/ -run Integration` | Run only integration tests (auth, products, orders). Use `go test`, not `go run`, for `*_test.go` files. |
| `go build ./...` | Verify all packages compile (no binary produced). |
| `go build -o api ./cmd/` | Build the API binary as `./api` (for deployment). |
| `sqlc generate` | Regenerate Go code from SQL. Run after changing `queries.sql` or schema. |
| `cp .env.example .env` | Create `.env` from the template (do once). |
| `source .env` then `goose up` | Run DB migrations (or use `./scripts/setup-and-run.sh` which does this for you). |

### One-shot setup and run

| Command | What it does |
|--------|----------------|
| `./scripts/setup-and-run.sh` | Start Docker Postgres, run migrations, seed products, then start the API. Good for “run everything” in one go. |

---

## 5. Order of operations (development)

1. **Start the database:**  
   `docker compose up -d`

2. **Ensure `.env` exists:**  
   `cp .env.example .env` (edit if you use a different DB port/user).

3. **Apply migrations (first time or after new migrations):**  
   `source .env` (or export `GOOSE_DBSTRING`), then `goose up`

4. **Start the API:**  
   `go run ./cmd`

5. **Test:**  
   Open Postman, hit `http://localhost:8080/health`, `http://localhost:8080/products`, `http://localhost:8080/orders`.

**How to stop:** Ctrl+C stops the API. To stop the DB: `docker compose down` (keeps data) or `docker compose down -v` (removes data). See README section "How to stop the API and the database" for the full table.

---

## 6. If something is “missing” (checklist)

- **API says “failed to connect to database”**  
  - Is Docker running? `docker ps` → see `ecom-postgres`.  
  - Is `.env` correct? Port **15432**, database **ecom**, user/password **postgres**.  
  - See [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

- **I don’t see data in my DB client**  
  - Connect to **port 15432**. See [MIGRATIONS_README.md](MIGRATIONS_README.md) section 7.

- **Migrations / schema**  
  - [MIGRATIONS_README.md](MIGRATIONS_README.md) – what migrations do, what happens to data when you change tables.

- **How the API and DB fit together**  
  - [ARCHITECTURE.md](ARCHITECTURE.md) – request flow, SQLC, when to run what after changes.

- **Credentials and config**  
  - All in `.env`; see main [README.md](README.md) section “Using .env for all credentials”.

---

## 7. Quick reference: development vs “deploy in Docker”

| | Development (now) | Later (deploy in Docker) |
|-|-------------------|---------------------------|
| **Database** | Docker (this repo’s `docker-compose`) | Same image in Docker, or a managed DB (e.g. cloud). |
| **API** | On your machine (`go run ./cmd`) | Can run in a container too (add a service in `docker-compose` or use a Dockerfile). |
| **Connection string** | `.env`: `localhost:15432` for Docker DB | In production, host might be `postgres` (service name) and port `5432` inside the Docker network. |

You’re not missing anything for **development**: use Docker for the DB, run the API on your machine, and use the commands above. When you’re ready to put the API in Docker as well, you’ll add a second service and point the API’s DSN to that Postgres service name and port.
