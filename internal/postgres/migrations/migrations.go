package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var files embed.FS

func Run(ctx context.Context, db *sql.DB) error {
	provider, err := goose.NewProvider(goose.DialectPostgres, db, files)
	if err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	return nil
}
