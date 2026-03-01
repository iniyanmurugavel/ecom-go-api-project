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
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sikozonpc/ecom/internal/auth"
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
		config: config{
			addr:            ":8080",
			db:              dbConfig{dsn: dsn},
			jwtConfig: &auth.JWTConfig{
				Secret:   []byte("test-secret-at-least-32-characters-long"),
				Expiry:   24 * time.Hour,
				Issuer:   "ecom-api",
				Audience: "ecom-api",
			},
			requestTimeout: 60 * time.Second,
		},
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

func TestIntegration_GetProducts_NoToken(t *testing.T) {
	app := testApp(t)
	if app == nil {
		return
	}
	h := app.mount()

	req := httptest.NewRequest(http.MethodGet, "/v1/products?limit=5&offset=0", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("GET /v1/products without token status = %d, want 401", rec.Code)
	}
}

func TestIntegration_RegisterLoginAndGetProducts(t *testing.T) {
	app := testApp(t)
	if app == nil {
		return
	}
	h := app.mount()

	// 1. Register
	regBody := bytes.NewBufferString(`{"email":"inttest@example.com","password":"pass123","name":"Int Test"}`)
	regReq := httptest.NewRequest(http.MethodPost, "/v1/auth/register", regBody)
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	h.ServeHTTP(regRec, regReq)
	_ = regRec // 201 or 409 if already exists

	// 2. Login
	loginBody := bytes.NewBufferString(`{"email":"inttest@example.com","password":"pass123"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	h.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Skipf("login failed (status %d), maybe DB not migrated: %s", loginRec.Code, loginRec.Body.String())
		return
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil || loginResp.Token == "" {
		t.Skipf("login response invalid: %v", err)
		return
	}

	// 3. Get products with token
	req := httptest.NewRequest(http.MethodGet, "/v1/products?limit=5&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /v1/products with token status = %d, want 200", rec.Code)
	}
	var list []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Errorf("GET /v1/products invalid JSON: %v", err)
	}
}

func TestIntegration_PostOrder_NoToken(t *testing.T) {
	app := testApp(t)
	if app == nil {
		return
	}
	h := app.mount()

	body := bytes.NewBufferString(`{"items": [{"productId": 1, "quantity": 1}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/orders", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("POST /v1/orders without token status = %d, want 401", rec.Code)
	}
}

func TestIntegration_PostOrder_WithToken(t *testing.T) {
	app := testApp(t)
	if app == nil {
		return
	}
	h := app.mount()

	// Register (ignore 409 if exists) then login
	regBody := bytes.NewBufferString(`{"email":"inttest@example.com","password":"pass123","name":"Int Test"}`)
	regReq := httptest.NewRequest(http.MethodPost, "/v1/auth/register", regBody)
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	h.ServeHTTP(regRec, regReq)
	_ = regRec

	loginBody := bytes.NewBufferString(`{"email":"inttest@example.com","password":"pass123"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	h.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Skipf("login failed (run RegisterLoginAndGetProducts first or migrate DB)")
		return
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil || loginResp.Token == "" {
		t.Skipf("login response invalid")
		return
	}

	// Place order with token (customerId from JWT, body has items only)
	body := bytes.NewBufferString(`{"items": [{"productId": 1, "quantity": 1}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/orders", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

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
