// Package main is the entry point: loads .env, connects to the DB with a connection pool,
// builds the router, and starts the HTTP server with graceful shutdown.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sikozonpc/ecom/internal/env"
)

func main() {
	// Load .env so getDSN() and env.GetString() can read DB and HTTP config.
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	// Logger used by handlers and services (and main). Structured logs go to stdout.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	jwtSecret := env.GetString("JWT_SECRET", "change-me-in-production-use-long-secret")
	if len(jwtSecret) < 32 {
		logger.Warn("JWT_SECRET should be at least 32 chars for security; using default for dev")
	}
	cfg := config{
		addr:      env.GetString("HTTP_ADDR", ":8080"),
		db:        dbConfig{dsn: getDSN()},
		rateLimit: env.GetInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100), // 0 = disabled
		jwtSecret: []byte(jwtSecret),
	}

	// Connection pool: many goroutines can use DB at once. Better than a single connection for scalability.
	pool, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		logger.Error("failed to connect to database", "error", err, "host", env.GetString("DB_HOST", "localhost"), "port", env.GetString("DB_PORT", "15432"), "db", env.GetString("DB_NAME", "ecom"))
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("connected to database", "host", env.GetString("DB_HOST", "localhost"), "port", env.GetString("DB_PORT", "15432"), "db", env.GetString("DB_NAME", "ecom"))

	app := application{config: cfg, pool: pool, logger: logger}
	handler := app.mount()

	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	// Start server in a goroutine so we can wait for shutdown signal in main.
	go func() {
		logger.Info("server started", "addr", cfg.addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown: on SIGINT (Ctrl+C) or SIGTERM, stop accepting new requests and let in-flight ones finish.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}
	logger.Info("server stopped")
}

// getDSN returns the Postgres connection string. Prefer GOOSE_DBSTRING from .env; else build from DB_* vars.
// Port 15432 = Docker Postgres for this project (5432 is often another Postgres on the host).
func getDSN() string {
	if dsn := os.Getenv("GOOSE_DBSTRING"); dsn != "" {
		if !strings.Contains(dsn, "port=") {
			dsn = dsn + " port=15432"
		} else if strings.Contains(dsn, "port=5432") {
			dsn = strings.Replace(dsn, "port=5432", "port=15432", 1)
		}
		return dsn
	}
	host := env.GetString("DB_HOST", "localhost")
	port := env.GetString("DB_PORT", "15432")
	user := env.GetString("DB_USER", "postgres")
	pass := env.GetString("DB_PASSWORD", "postgres")
	dbname := env.GetString("DB_NAME", "ecom")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
}
