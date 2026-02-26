# Development and Deployment: Basics for Beginners

This file answers: **Should I use Docker or my own DB in development?** and **Do I deploy in Docker later?** It also explains Docker basics and gives you a simple command reference. If you're new to Docker or deployment, read this once and keep it as reference.

---

## 1. Development on your machine: Docker for DB, API on your machine (recommended)

### What this project does today

| What | Where it runs in development |
|------|------------------------------|
| **PostgreSQL (database)** | Inside **Docker** (container `ecom-postgres`). You start it with `docker compose up -d`. |
| **Your Go API** | On **your machine** (not in Docker). You run `go run cmd/*.go` in the terminal. |

So in dev:

- **Database** → Docker (so you don’t install Postgres on your Mac and avoid port clashes with other tools).
- **API** → Your machine (easy to edit code, run tests, use debugger).

You **do not** need to install PostgreSQL on your Mac for this project. Docker gives you a Postgres that’s isolated and matches the same setup you can use later (e.g. in production or in a full Docker deploy).

### Can I use my own PostgreSQL (installed on my Mac) instead of Docker?

Yes. If you already have Postgres (e.g. Postgres.app, Homebrew) and want to use it:

1. Create a database named `ecom` and a user `postgres` with password `postgres` (or whatever you prefer).
2. In `.env`, set `GOOSE_DBSTRING` (and optionally `DB_*`) to point to **your** Postgres (usually `port=5432`).
3. Run migrations: `goose up`.
4. Run the API: `go run cmd/*.go`.

The app doesn’t care whether the DB is in Docker or on your host; it only cares about the connection string in `.env`. Using Docker is recommended so everyone has the same setup and you avoid “works on my machine” issues.

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

- **Development:** Use **Docker for the DB** (recommended); run the **API on your machine** with `go run cmd/*.go`.  
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
| `go run cmd/*.go` | Build and run the API. It reads `.env` and connects to Postgres (port 15432). Stop with Ctrl+C. |
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
   `go run cmd/*.go`

5. **Test:**  
   Open Postman, hit `http://localhost:8080/health`, `http://localhost:8080/products`, `http://localhost:8080/orders`.

To stop: Ctrl+C for the API; `docker compose down` if you want to stop Postgres.

---

## 6. If something is “missing” (checklist)

- **API says “failed to connect to database”**  
  - Is Docker running? `docker ps` → see `ecom-postgres`.  
  - Is `.env` correct? Port **15432**, database **ecom**, user/password **postgres**.  
  - See [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

- **I don’t see data in my DB client**  
  - Connect to **port 15432** (not 5432). See [MIGRATIONS_README.md](MIGRATIONS_README.md) section 7.

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
| **API** | On your machine (`go run cmd/*.go`) | Can run in a container too (add a service in `docker-compose` or use a Dockerfile). |
| **Connection string** | `.env`: `localhost:15432` for Docker DB | In production, host might be `postgres` (service name) and port `5432` inside the Docker network. |

You’re not missing anything for **development**: use Docker for the DB, run the API on your machine, and use the commands above. When you’re ready to put the API in Docker as well, you’ll add a second service and point the API’s DSN to that Postgres service name and port.
