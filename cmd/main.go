// Package main is the entry point: loads .env, connects to DB, builds router, starts HTTP server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/sikozonpc/ecom/internal/env"
)

func main() {
	// .env is loaded into os.Environ(); getDSN() and env.GetString() read from there.
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	ctx := context.Background()
	cfg := config{
		addr: env.GetString("HTTP_ADDR", ":8080"),
		db:   dbConfig{dsn: getDSN()},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Single DB connection used by all handlers (via repo.Queries).
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		logger.Error("failed to connect to database", "error", err, "host", env.GetString("DB_HOST", "localhost"), "port", env.GetString("DB_PORT", "15432"), "db", env.GetString("DB_NAME", "ecom"))
		os.Exit(1)
	}
	defer conn.Close(ctx)

	logger.Info("connected to database", "host", env.GetString("DB_HOST", "localhost"), "port", env.GetString("DB_PORT", "15432"), "db", env.GetString("DB_NAME", "ecom"))

	// mount() builds chi router + middleware and wires handlers; run() starts http.Server.
	api := application{config: cfg, db: conn}
	if err := api.run(api.mount()); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}

// getDSN returns the Postgres DSN. Prefer GOOSE_DBSTRING from .env; else build from DB_* vars. Port 15432 = Docker (5432 = other Postgres, often no "postgres" role).
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
