// Integration tests: hit real HTTP handlers with a real DB (router + handler + service + repo + Postgres).
// Requires a running Postgres (e.g. docker compose up -d) and .env or GOOSE_DBSTRING. Skips if DB is unavailable.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
}

// testApp builds an application with a real pool for integration tests. Returns nil if DB is unavailable.
// Caller should use the returned handler in the same test; pool is closed when the test ends.
func testApp(t *testing.T) *application {
	t.Helper()
	dsn := getDSN()
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Skipf("integration test skipped: no DB: %v", err)
		return nil
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("integration test skipped: DB ping failed: %v", err)
		return nil
	}
	t.Cleanup(func() { pool.Close() })
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return &application{
		config: config{addr: ":8080", db: dbConfig{dsn: dsn}},
		pool:   pool,
		logger: logger,
	}
}

func TestIntegration_Health(t *testing.T) {
	app := testApp(t)
	if app == nil {
		return
	}
	h := app.mount()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /health status = %d, want 200", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "all good" {
		t.Errorf("GET /health body = %q, want \"all good\"", body)
	}
}

func TestIntegration_HealthLive(t *testing.T) {
	app := testApp(t)
	if app == nil {
		return
	}
	h := app.mount()

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /health/live status = %d, want 200 (DB must be up)", rec.Code)
	}
}

func TestIntegration_GetProducts(t *testing.T) {
	app := testApp(t)
	if app == nil {
		return
	}
	h := app.mount()

	req := httptest.NewRequest(http.MethodGet, "/v1/products?limit=5&offset=0", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /v1/products status = %d, want 200", rec.Code)
	}
	// Response should be JSON array (possibly empty)
	var list []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Errorf("GET /v1/products invalid JSON: %v", err)
	}
}

func TestIntegration_PostOrder_InvalidBody(t *testing.T) {
	app := testApp(t)
	if app == nil {
		return
	}
	h := app.mount()

	body := bytes.NewBufferString(`{"customerId": 1}`) // no "items"
	req := httptest.NewRequest(http.MethodPost, "/v1/orders", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /v1/orders (no items) status = %d, want 400", rec.Code)
	}
}

func TestIntegration_PostOrder_ValidBody(t *testing.T) {
	app := testApp(t)
	if app == nil {
		return
	}
	h := app.mount()

	// Assume product id 1 exists (seed or migration). If not, we may get 404.
	body := bytes.NewBufferString(`{"customerId": 1, "items": [{"productId": 1, "quantity": 1}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/orders", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// 201 created or 404 if product 1 does not exist
	if rec.Code != http.StatusCreated && rec.Code != http.StatusNotFound {
		t.Errorf("POST /v1/orders status = %d, want 201 or 404", rec.Code)
	}
	if rec.Code == http.StatusCreated {
		var order map[string]interface{}
		if err := json.NewDecoder(rec.Body).Decode(&order); err != nil {
			t.Errorf("POST /v1/orders response JSON: %v", err)
		}
		if _, ok := order["id"]; !ok {
			t.Errorf("POST /v1/orders response missing id")
		}
	}
}
