// API wiring: chi router, middleware, and route -> handler -> service -> repo.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	repo "github.com/sikozonpc/ecom/internal/adapters/postgresql/sqlc"
	"github.com/sikozonpc/ecom/internal/orders"
	"github.com/sikozonpc/ecom/internal/products"
)

// mount builds the HTTP handler: middleware stack then route -> handler. Handler gets service, service gets repo (DB).
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// Middleware runs on every request before it reaches your handler. Order matters.
	r.Use(middleware.RequestID) // gives each request a unique ID (useful for logs and rate limiting)
	r.Use(middleware.RealIP)    // reads the real client IP from headers (e.g. behind a proxy)
	r.Use(middleware.Logger)    // logs each request (method, path, status, duration)
	r.Use(middleware.Recoverer) // if a handler panics, this catches it and returns 500 instead of crashing the server

	// If a request takes longer than 60 seconds, cancel it. ctx.Done() will signal so your code can stop.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good"))
	})

	// Dependency chain: handler -> service -> repo. Repo = sqlc-generated code that runs SQL.
	productService := products.NewService(repo.New(app.db))
	r.Get("/products", products.NewHandler(productService).ListProducts)

	// Orders need db connection for transactions; repo.New(app.db) for queries, app.db for Begin/Commit.
	orderService := orders.NewService(repo.New(app.db), app.db)
	r.Post("/orders", orders.NewHandler(orderService).PlaceOrder)

	return r
}

func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	log.Printf("server has started at addr %s", app.config.addr)
	return srv.ListenAndServe()
}

// application holds config and the single DB connection passed into repos.
type application struct {
	config config
	db     *pgx.Conn
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}
