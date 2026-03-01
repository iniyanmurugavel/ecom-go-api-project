// Package main is the entry point: loads .env, connects to the DB with a connection pool,
// builds the router, and starts the HTTP server with graceful shutdown.
//
// LEARNING: In Go, the main package is special — it must have a func main() that runs first.
// Other packages live under internal/ and are imported by main.
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
	// LEARNING: godotenv loads key=value from .env into os.Getenv(). We ignore the error
	// because .env might not exist (e.g. in CI where env vars are set differently).
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	// LEARNING: slog is Go's structured logger (Go 1.21+). TextHandler writes to stdout.
	// slog.SetDefault makes this the logger used by packages that call slog.Info/Error.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	jwtSecret := env.GetString("JWT_SECRET", "change-me-in-production-use-long-secret")
	if len(jwtSecret) < 32 {
		logger.Warn("JWT_SECRET should be at least 32 chars for security; using default for dev")
	}
	// LEARNING: Parse CORS_ALLOWED_ORIGINS (comma-separated origins for browser requests).
	// Empty = CORS disabled. Mobile apps don't need CORS (only browsers enforce it).
	corsOrigins := []string{}
	if s := env.GetString("CORS_ALLOWED_ORIGINS", ""); s != "" {
		for _, o := range strings.Split(s, ",") {
			if o = strings.TrimSpace(o); o != "" {
				corsOrigins = append(corsOrigins, o)
			}
		}
	}
	cfg := config{
		addr:        env.GetString("HTTP_ADDR", ":8080"),
		db:          dbConfig{dsn: getDSN()},
		rateLimit:   env.GetInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100), // 0 = disabled
		jwtSecret:   []byte(jwtSecret),
		corsOrigins: corsOrigins,
	}

	// Connection pool: many goroutines can use DB at once. Tune via DB_POOL_MAX_CONNS, DB_POOL_MIN_CONNS.
	// LEARNING: A pool reuses connections instead of opening one per request. MaxConns caps total;
	// MinConns keeps idle connections ready for faster requests.
	poolConfig, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil {
		logger.Error("failed to parse database config", "error", err)
		os.Exit(1)
	}
	poolConfig.MaxConns = int32(env.GetInt("DB_POOL_MAX_CONNS", 25))
	poolConfig.MinConns = int32(env.GetInt("DB_POOL_MIN_CONNS", 2))
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("failed to connect to database", "error", err, "host", env.GetString("DB_HOST", "localhost"), "port", env.GetString("DB_PORT", "15432"), "db", env.GetString("DB_NAME", "ecom"))
		os.Exit(1)
	}
	// LEARNING: defer runs the function when the surrounding function returns (e.g. on exit).
	// So pool.Close() is always called when main exits, releasing DB connections.
	defer pool.Close()

	logger.Info("connected to database", "host", env.GetString("DB_HOST", "localhost"), "port", env.GetString("DB_PORT", "15432"), "db", env.GetString("DB_NAME", "ecom"))

	// LEARNING: application is a struct that holds config, pool, and logger. We pass it to mount()
	// so routes can access the DB and config. See cmd/api.go for the application type.
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
	// LEARNING: "go func()" starts a new goroutine (lightweight thread). ListenAndServe blocks,
	// so we run it in the background. Main goroutine then blocks on <-quit until SIGINT/SIGTERM.
	go func() {
		logger.Info("server started", "addr", cfg.addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown: on SIGINT (Ctrl+C) or SIGTERM, stop accepting new requests and let in-flight ones finish.
	// LEARNING: Channels are Go's way to pass data between goroutines. make(chan os.Signal, 1)
	// creates a buffered channel. signal.Notify sends SIGINT/SIGTERM to quit. <-quit blocks
	// until a signal is received. Then we call srv.Shutdown to stop the server gracefully.
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

// getDSN returns the Postgres connection string. Uses GOOSE_DBSTRING from .env if set;
// otherwise builds from DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME.
// Local: port 15432 (Docker). CI: port 5432 (GitHub Actions).
func getDSN() string {
	if dsn := os.Getenv("GOOSE_DBSTRING"); dsn != "" {
		return dsn
	}
	host := env.GetString("DB_HOST", "localhost")
	port := env.GetString("DB_PORT", "15432")
	user := env.GetString("DB_USER", "postgres")
	pass := env.GetString("DB_PASSWORD", "postgres")
	dbname := env.GetString("DB_NAME", "ecom")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
}
