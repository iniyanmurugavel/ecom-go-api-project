// API wiring: chi router, middleware, routes under /v1, and dependency chain handler -> service -> repo.
// Repo is sqlc-generated; services depend on interfaces so we can test with mocks.
package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
	"github.com/sikozonpc/ecom/internal/auth"
	"github.com/sikozonpc/ecom/internal/orders"
	"github.com/sikozonpc/ecom/internal/products"
)

// mount builds the HTTP handler: middleware, then routes. Health at /health; API under /v1.
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// Middleware runs on every request before your handler. Order matters.
	r.Use(middleware.RequestID)   // unique ID per request (logs, tracing)
	r.Use(middleware.RealIP)      // client IP from X-Forwarded-For when behind a proxy
	r.Use(middleware.Logger)      // log method, path, status, duration
	r.Use(middleware.Recoverer)   // catch panics and return 500 instead of crashing
	r.Use(middleware.Timeout(60 * time.Second)) // cancel long-running requests after 60s

	// Rate limit per IP (production). Set app.config.rateLimit to 0 to disable (e.g. tests).
	if app.config.rateLimit > 0 {
		r.Use(httprate.LimitByIP(app.config.rateLimit, time.Minute))
	}

	// Health: simple 200 for load balancers. /health/live pings the DB (503 if DB down).
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("all good"))
	})
	r.Get("/health/live", app.healthLive)

	// API v1: Auth (public), Products and Orders (JWT protected).
	r.Route("/v1", func(r chi.Router) {
		// Auth: register and login (no token required).
		authSvc := auth.NewService(repo.New(app.pool))
		authHandler := auth.NewHandler(authSvc, app.config.jwtSecret)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Protected routes: require Authorization: Bearer <token>. Customer ID from JWT.
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(app.config.jwtSecret))
			productRepo := repo.New(app.pool)
			productSvc := products.NewService(productRepo, app.logger)
			productHandler := products.NewHandler(productSvc, app.logger)
			r.Get("/products", productHandler.ListProducts)

			orderSvc := orders.NewService(orders.NewOrderRepo(repo.New(app.pool)), app.pool, app.logger)
			orderHandler := orders.NewHandler(orderSvc, app.logger)
			r.Post("/orders", orderHandler.PlaceOrder)
		})
	})

	return r
}

// healthLive pings the DB. Returns 200 if OK, 503 if DB is down (so load balancers can remove the instance).
func (app *application) healthLive(w http.ResponseWriter, r *http.Request) {
	if err := app.pool.Ping(r.Context()); err != nil {
		app.logger.Error("health live: db ping failed", "error", err)
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Write([]byte("ok"))
}

// application holds config, the DB pool, and logger. Passed to mount and healthLive.
type application struct {
	config config
	pool   *pgxpool.Pool
	logger *slog.Logger
}

type config struct {
	addr      string
	db        dbConfig
	rateLimit int    // requests per minute per IP; 0 = disabled
	jwtSecret []byte // for signing/verifying JWTs
}

type dbConfig struct {
	dsn string
}
