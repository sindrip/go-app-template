package integration_test

import (
	"database/sql"
	"fmt"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/sindrip/go-app-template/internal/app"
	"github.com/sindrip/go-app-template/internal/postgres"
	"github.com/stretchr/testify/require"
)

var databaseCount atomic.Int64

func newDatabase(t *testing.T) string {
	t.Helper()

	admin, err := sql.Open("pgx", postgresURL)
	require.NoError(t, err)

	defer func() { require.NoError(t, admin.Close()) }()

	name := fmt.Sprintf("test_%d", databaseCount.Add(1))

	_, err = admin.ExecContext(t.Context(), "CREATE DATABASE "+name)
	require.NoError(t, err)

	databaseURL, err := url.Parse(postgresURL)
	require.NoError(t, err)

	databaseURL.Path = "/" + name

	db, err := sql.Open("pgx", databaseURL.String())
	require.NoError(t, err)

	defer func() { require.NoError(t, db.Close()) }()

	_, err = db.ExecContext(t.Context(), postgres.Schema)
	require.NoError(t, err)

	return databaseURL.String()
}

func startApp(t *testing.T, databaseURL string) *app.App {
	t.Helper()

	ctx := t.Context()

	env := map[string]string{
		"PORT":         "0",
		"PROBE_PORT":   "0",
		"DATABASE_URL": databaseURL,
	}

	a, err := app.New(ctx, func(key string) (string, bool) {
		value, ok := env[key]

		return value, ok
	}, t.Output())
	require.NoError(t, err)

	errc := make(chan error, 1)

	go func() { errc <- a.Run(ctx) }()

	t.Cleanup(func() { require.NoError(t, <-errc) })

	return a
}

func serverURL(a *app.App, path string) string {
	return "http://" + a.Addr().String() + path
}

func probeURL(a *app.App, path string) string {
	return "http://" + a.ProbeAddr().String() + path
}
