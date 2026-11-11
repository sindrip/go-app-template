package lifecycle

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
)

func (l *Lifecycle) Listen(ctx context.Context, port uint16) (net.Listener, error) {
	var lc net.ListenConfig

	ln, err := lc.Listen(ctx, "tcp", net.JoinHostPort("", strconv.Itoa(int(port))))
	if err != nil {
		return nil, err
	}

	l.cleanups = append(l.cleanups, func(context.Context) error {
		if err := ln.Close(); !errors.Is(err, net.ErrClosed) {
			return err
		}

		return nil
	})

	return ln, nil
}

type ListenServer interface {
	Serve(ln net.Listener) error
	Shutdown(ctx context.Context) error
}

func (l *Lifecycle) Serve[T ListenServer](ln net.Listener, srv T) T {
	l.runs = append(l.runs, func() error {
		if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	l.cleanups = append(l.cleanups, srv.Shutdown)

	return srv
}
