# Docs Folder — Learning & API Reference

All learning docs and API reference for this project are in this folder. **Start with [LEARNING_PATH.md](LEARNING_PATH.md)** for a stack-ranked order to learn everything.

---

## Quick Index

### Learning

| Doc | What it covers |
|-----|-----------------|
| **[LEARNING_PATH.md](LEARNING_PATH.md)** | **Start here.** Step-by-step order to learn the project (1 → 13). |
| **[CORS_AND_CLIENTS.md](CORS_AND_CLIENTS.md)** | CORS explained with examples. Why web needs it, why mobile doesn't. |
| **[LEARNER_GUIDE.md](LEARNER_GUIDE.md)** | Project structure, request flow diagram, Go concepts. |
| **[MAC_POSTGRES_15432.md](MAC_POSTGRES_15432.md)** | Use Mac Postgres on port 15432 instead of Docker. |
| **[../DEPLOYMENT.md](../DEPLOYMENT.md)** | Production deployment (with/without Docker), where to deploy, checklist. |

### API & Testing

| Doc | What it covers |
|-----|-----------------|
| **[../RUN_AND_TEST.md](../RUN_AND_TEST.md)** | How to run the API and run unit + integration tests. |
| **[API_REFERENCE.md](API_REFERENCE.md)** | **All endpoints** — curl commands + expected responses. Single source of truth. |
| **[POSTMAN_CLI.md](POSTMAN_CLI.md)** | Run Postman collection from terminal (Newman). |
| **[API_DOCUMENTATION.md](API_DOCUMENTATION.md)** | OpenAPI/Swagger — view spec, generate clients. |
| **[openapi.yaml](openapi.yaml)** | OpenAPI 3 spec for tooling (Swagger UI, code gen). |

### Database & Migrations

| Doc | What it covers |
|-----|-----------------|
| **[DATABASE_AND_TABLES_STRATEGY.md](DATABASE_AND_TABLES_STRATEGY.md)** | Table design, relationships, why we structured the DB this way. |
| **[MIGRATION_EXAMPLE.md](MIGRATION_EXAMPLE.md)** | Step-by-step: add a column without data loss. |
| **[ACID_EXPLAINED.md](ACID_EXPLAINED.md)** | ACID properties (Atomicity, Consistency, Isolation, Durability) with examples. |
| **[REDIS_GUIDE.md](REDIS_GUIDE.md)** | Why Redis, when to use it (caching, rate limit, JWT blocklist), how to add it. |

### Deployment

| Doc | What it covers |
|-----|-----------------|
| **[NGINX_GUIDE.md](NGINX_GUIDE.md)** | Why nginx is needed, when to use it, example configs (HTTPS, reverse proxy). |

---

## Suggested Order

1. [LEARNING_PATH.md](LEARNING_PATH.md) — follow steps 1–13 in order
2. [API_REFERENCE.md](API_REFERENCE.md) — when testing endpoints (curl, Postman)
3. [POSTMAN_CLI.md](POSTMAN_CLI.md) — when running tests from terminal
4. [API_DOCUMENTATION.md](API_DOCUMENTATION.md) — when using Swagger/OpenAPI tools
