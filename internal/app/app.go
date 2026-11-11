package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sindrip/go-app-template/internal/server"
	"github.com/sindrip/go-app-template/pkg/config"
	"github.com/sindrip/go-app-template/pkg/lifecycle"
	"github.com/sindrip/go-app-template/pkg/probes"
)

const readHeaderTimeout = 10 * time.Second

type App struct {
	cfg       Config
	logger    *slog.Logger
	lifecycle *lifecycle.Lifecycle

	probes net.Listener
	server net.Listener
}

type Config struct {
	Port                   uint16 `env:"PORT"`
	ProbePort              uint16 `env:"PROBE_PORT"`
	ShutdownTimeoutSeconds uint   `env:"SHUTDOWN_TIMEOUT_SECONDS"`
	DatabaseURL            string `env:"DATABASE_URL"`
}

func New(ctx context.Context, env config.Lookup, stdout io.Writer) (*App, error) {
	cfg, err := env.Defaults(Config{
		ShutdownTimeoutSeconds: 25,
	}).Load()
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(stdout, nil))
	lc := lifecycle.New()

	a, err := build(ctx, cfg, logger, lc)
	if err != nil {
		return nil, errors.Join(err, lc.Shutdown(ctx))
	}

	return a, nil
}

func build(
	ctx context.Context,
	cfg Config,
	logger *slog.Logger,
	lc *lifecycle.Lifecycle,
) (*App, error) {
	probeListener, err := lc.Listen(ctx, cfg.ProbePort)
	if err != nil {
		return nil, fmt.Errorf("listen probes: %w", err)
	}

	lc.Serve(probeListener, &http.Server{
		Handler:           probes.Handler(probes.Checks{Ready: lc.Ready}),
		ReadHeaderTimeout: readHeaderTimeout,
	})

	db, err := lc.Database(sql.Open("pgx", cfg.DatabaseURL))
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	serverListener, err := lc.Listen(ctx, cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	lc.Serve(serverListener, &http.Server{
		Handler:           server.New(db),
		ReadHeaderTimeout: readHeaderTimeout,
	})

	return &App{
		cfg:       cfg,
		logger:    logger,
		lifecycle: lc,
		probes:    probeListener,
		server:    serverListener,
	}, nil
}

func (a *App) Addr() net.Addr {
	return a.server.Addr()
}

func (a *App) ProbeAddr() net.Addr {
	return a.probes.Addr()
}

func (a *App) Run(ctx context.Context) error {
	timeout := time.Duration(a.cfg.ShutdownTimeoutSeconds) * time.Second

	a.logger.InfoContext(ctx, "serving",
		"addr", a.server.Addr().String(),
		"probe_addr", a.probes.Addr().String(),
	)

	return a.lifecycle.Run(ctx, timeout)
}
