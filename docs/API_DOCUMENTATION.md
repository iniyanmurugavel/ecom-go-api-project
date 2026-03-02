# API Documentation — Swagger / OpenAPI

This project includes an **OpenAPI 3 spec** for tooling integration (Swagger UI, Redoc, code generation). We do **not** serve Swagger UI from the API itself — you view the spec with external tools.

---

## 1. What We Have

| Item | Location | Purpose |
|------|----------|---------|
| **OpenAPI spec** | [docs/openapi.yaml](openapi.yaml) | Machine-readable API definition |
| **Human reference** | [docs/API_REFERENCE.md](API_REFERENCE.md) | curl + responses for every endpoint |

---

## 2. View the OpenAPI Spec

### Option A: Swagger Editor (online)

1. Go to [editor.swagger.io](https://editor.swagger.io)
2. File → Import file → select `docs/openapi.yaml`
3. View and edit the spec

### Option B: Swagger UI (Docker)

From project root:

```bash
docker run -p 8081:8080 -e SWAGGER_JSON_URL=/openapi.yaml \
  -v $(pwd)/docs/openapi.yaml:/openapi.yaml \
  swaggerapi/swagger-ui
```

Open `http://localhost:8081` — the spec loads automatically.

### Option C: Redoc (Docker)

```bash
docker run -p 8081:80 -v $(pwd)/docs/openapi.yaml:/usr/share/nginx/html/openapi.yaml \
  redocly/redoc
```

Then open `http://localhost:8081` and navigate to `/openapi.yaml`.

### Option D: VS Code extension

Install **OpenAPI (Swagger) Editor** or **Swagger Viewer**. Open `docs/openapi.yaml` and use the preview.

---

## 3. Use the Spec for Code Generation

### Generate client (e.g. TypeScript)

```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/openapi.yaml \
  -g typescript-fetch \
  -o ./generated-client
```

### Generate server stub (e.g. Go)

```bash
npx @openapitools/openapi-generator-cli generate \
  -i docs/openapi.yaml \
  -g go-server \
  -o ./generated-server
```

---

## 4. Serve Swagger UI from the API (optional)

If you want Swagger UI at `/docs` served by the Go API:

1. Add a static file server or embed the Swagger UI assets
2. Serve `docs/openapi.yaml` at `/openapi.yaml`
3. Configure Swagger UI to load it

Example with `github.com/swaggo/http-swagger`:

```go
// Add to your router
http.Handle("/docs/", httpSwagger.Handler(
    httpSwagger.URL("/openapi.yaml"),
))
```

You would need to add the swaggo packages and wire this in. The current project keeps docs separate for simplicity.

---

## 5. Keeping the Spec Updated

When you add or change endpoints:

1. Edit `docs/openapi.yaml`
2. Update `docs/API_REFERENCE.md` (curl + responses)
3. Re-import into Swagger Editor to validate

---

## 6. Related Docs

- [API_REFERENCE.md](API_REFERENCE.md) — All curl commands and responses
- `scripts/verify-api.sh` — Automated verification of all 7 endpoints (run with API up)
- [POSTMAN_CLI.md](POSTMAN_CLI.md) — Run Postman collection from CLI
