// API wiring: chi router, middleware, routes under /v1, and dependency chain handler -> service -> repo.
// Repo is sqlc-generated; services depend on interfaces so we can test with mocks.
//
// LEARNING: This file defines the HTTP layer. Chi is a lightweight router — it matches URLs
// to handlers. Middleware runs before handlers (logging, auth, rate limit, etc.).
package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
	"github.com/sikozonpc/ecom/internal/auth"
	"github.com/sikozonpc/ecom/internal/json"
	"github.com/sikozonpc/ecom/internal/metrics"
	"github.com/sikozonpc/ecom/internal/orders"
	"github.com/sikozonpc/ecom/internal/products"
)

// mount builds the HTTP handler: middleware, then routes. Health at /health; API under /v1.
// LEARNING: In Go, (app *application) is a "method receiver" — mount belongs to application.
// It returns http.Handler, which is the interface that ServeHTTP(w, r). Chi's router implements it.
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// Middleware runs on every request before your handler. Order matters.
	// LEARNING: r.Use() adds middleware. Each middleware can wrap the next handler or short-circuit.
	requestTimeout := app.config.requestTimeout
	r.Use(middleware.RequestID)   // unique ID per request (logs, tracing)
	r.Use(middleware.RealIP)      // client IP from X-Forwarded-For when behind a proxy
	r.Use(middleware.Logger)      // log method, path, status, duration
	r.Use(middleware.Recoverer)   // catch panics and return 500 instead of crashing
	r.Use(middleware.Timeout(requestTimeout)) // cancel long-running requests
	r.Use(metrics.Middleware)     // Prometheus request count and duration

	// Rate limit per IP (production). Set app.config.rateLimit to 0 to disable (e.g. tests).
	if app.config.rateLimit > 0 {
		r.Use(httprate.LimitByIP(app.config.rateLimit, time.Minute))
	}

	// CORS: allow browser requests from configured origins. Comma-separated list in CORS_ALLOWED_ORIGINS.
	if len(app.config.corsOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   app.config.corsOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			AllowCredentials: true,
		}))
	}

	// Health: JSON 200 for load balancers. /health/live pings the DB (503 if DB down).
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		json.Write(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/health/live", app.healthLive)
	r.Handle("/metrics", metrics.Handler())

	// API v1: Auth (public), Products and Orders (JWT protected).
	r.Route("/v1", func(r chi.Router) {
		// Auth: register and login (no token required).
		authSvc := auth.NewService(auth.NewAuthRepo(repo.New(app.pool)))
		authHandler := auth.NewHandler(authSvc, app.config.jwtConfig, app.logger)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Protected routes: require Authorization: Bearer <token>. Customer ID from JWT.
		// LEARNING: r.Group() creates a sub-router. All routes inside get the RequireAuth middleware.
		// RequireAuth runs before the handler — it checks the JWT and puts customer ID in context.
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(app.config.jwtConfig))
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
		json.WriteError(w, r, http.StatusServiceUnavailable, "database unavailable")
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
	addr            string
	db              dbConfig
	rateLimit       int               // requests per minute per IP; 0 = disabled
	jwtConfig       *auth.JWTConfig   // JWT secret, expiry, issuer, audience
	corsOrigins     []string          // CORS allowed origins; empty = CORS disabled
	requestTimeout  time.Duration    // max time per request (middleware)
}

type dbConfig struct {
	dsn string
}
