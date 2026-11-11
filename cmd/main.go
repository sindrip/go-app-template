package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sindrip/go-app-template/internal/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	context.AfterFunc(ctx, stop)

	a, err := app.New(ctx, os.LookupEnv, os.Stdout)
	if err != nil {
		return err
	}

	return a.Run(ctx)
}
