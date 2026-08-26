package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/pkg/config"
	"github.com/prepyo/backend/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Env)
	log.Info("running database migrations", "database_url", maskURL(cfg.DatabaseURL))

	ctx := context.Background()

	// 1. Ensure the target database exists (connect to default postgres db if needed)
	if err := ensureDatabaseExists(ctx, cfg.DatabaseURL, log); err != nil {
		log.Warn("ensure database exists check completed with note", "detail", err)
	}

	// 2. Connect to the target prepyo database
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 3. Run all migrations
	if err := database.Migrate(ctx, pool, log); err != nil {
		log.Error("migration failed", "error", err)
		os.Exit(1)
	}

	log.Info("✅ All migrations applied successfully to PostgreSQL database!")
}

func ensureDatabaseExists(ctx context.Context, dbURL string, log *slog.Logger) error {
	u, err := url.Parse(dbURL)
	if err != nil {
		return err
	}

	targetDB := strings.TrimPrefix(u.Path, "/")
	if targetDB == "" || targetDB == "postgres" {
		return nil
	}

	// Connect to default 'postgres' maintenance database
	adminURL := *u
	adminURL.Path = "/postgres"

	conn, err := pgx.Connect(ctx, adminURL.String())
	if err != nil {
		return fmt.Errorf("connect to maintenance db: %w", err)
	}
	defer conn.Close(ctx)

	var exists bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", targetDB).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check db existence: %w", err)
	}

	if !exists {
		log.Info("database does not exist, creating database", "dbname", targetDB)
		// CREATE DATABASE cannot run inside a transaction block or with parameters
		_, err = conn.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, strings.ReplaceAll(targetDB, `"`, `""`)))
		if err != nil {
			return fmt.Errorf("create database %s: %w", targetDB, err)
		}
		log.Info("created database successfully", "dbname", targetDB)
	} else {
		log.Info("database already exists", "dbname", targetDB)
	}

	return nil
}

func maskURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "invalid-url"
	}
	return u.Redacted()
}
