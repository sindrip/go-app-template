package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	postgresImage    = "postgres:18-alpine"
	postgresUser     = "postgres"
	postgresPassword = "secret"

	ownerUser = "app_owner"
	appUser   = "app"

	createRoles = `
CREATE ROLE app_readwrite NOLOGIN;
CREATE ROLE app_owner LOGIN PASSWORD 'secret';
CREATE ROLE app LOGIN PASSWORD 'secret';
GRANT app_readwrite TO app;
`
)

var postgresURL string

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()

	container, err := postgres.Run(ctx, postgresImage,
		postgres.WithUsername(postgresUser),
		postgres.WithPassword(postgresPassword),
		postgres.BasicWaitStrategies(),
	)

	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	postgresURL, err = container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	if err := provisionRoles(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	return m.Run()
}

func provisionRoles(ctx context.Context) error {
	admin, err := sql.Open("pgx", postgresURL)
	if err != nil {
		return err
	}

	defer func() { _ = admin.Close() }()

	_, err = admin.ExecContext(ctx, createRoles)

	return err
}
