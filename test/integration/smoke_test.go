package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const unreachableDatabaseURL = "postgres://app:secret@127.0.0.1:1/app"

func TestProbes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{name: "startup", path: "/startup"},
		{name: "live", path: "/live"},
		{name: "ready", path: "/ready"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := startApp(t, newDatabase(t))

			req, err := http.NewRequestWithContext(
				t.Context(),
				http.MethodGet,
				probeURL(a, tt.path),
				nil,
			)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())

			require.Equal(t, http.StatusOK, res.StatusCode)
		})
	}
}

func TestReadyWithoutDatabase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want int
	}{
		{name: "live", path: "/live", want: http.StatusOK},
		{name: "ready", path: "/ready", want: http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := startApp(t, unreachableDatabaseURL)

			req, err := http.NewRequestWithContext(
				t.Context(),
				http.MethodGet,
				probeURL(a, tt.path),
				nil,
			)
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())

			require.Equal(t, tt.want, res.StatusCode)
		})
	}
}
