package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sindrip/go-app-template/internal/postgres/migrations"
	"github.com/sindrip/go-app-template/pkg/config"
)

type Config struct {
	DatabaseURL string `env:"MIGRATION_DATABASE_URL"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Lookup(os.LookupEnv).Load[Config]()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if cfg.DatabaseURL == "" {
		return errors.New("config: MIGRATION_DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	defer func() { _ = db.Close() }()

	return migrations.Run(ctx, db)
}
