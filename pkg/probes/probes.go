package probes

import (
	"context"
	"net/http"
)

type Check func(ctx context.Context) error

type Checks struct {
	Startup Check
	Live    Check
	Ready   Check
}

func Handler(checks Checks) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /startup", probe(checks.Startup))
	mux.Handle("GET /live", probe(checks.Live))
	mux.Handle("GET /ready", probe(checks.Ready))

	return mux
}

func probe(check Check) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check == nil {
			return
		}

		if err := check(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
	})
}
