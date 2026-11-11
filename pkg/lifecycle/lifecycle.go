package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
)

type Lifecycle struct {
	runs     []func() error
	cleanups []func(context.Context) error
	checks   []check
}

type check struct {
	name string
	fn   func(context.Context) error
}

func New() *Lifecycle {
	return &Lifecycle{}
}

func (l *Lifecycle) Check(name string, fn func(ctx context.Context) error) {
	l.checks = append(l.checks, check{name: name, fn: fn})
}

func (l *Lifecycle) Ready(ctx context.Context) error {
	errs := make([]error, len(l.checks))

	var wg sync.WaitGroup

	for i, c := range l.checks {
		wg.Go(func() {
			if err := c.fn(ctx); err != nil {
				errs[i] = fmt.Errorf("%s: %w", c.name, err)
			}
		})
	}

	wg.Wait()

	return errors.Join(errs...)
}

func (l *Lifecycle) Run(ctx context.Context, timeout time.Duration) error {
	errc := make(chan error, len(l.runs))

	for _, run := range l.runs {
		go func() { errc <- run() }()
	}

	running := len(l.runs)

	var err error

	select {
	case err = <-errc:
		running--
	case <-ctx.Done():
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	errs := []error{err, l.Shutdown(ctx)}

	for range running {
		select {
		case err := <-errc:
			errs = append(errs, err)
		case <-ctx.Done():
			return errors.Join(append(errs, ctx.Err())...)
		}
	}

	return errors.Join(errs...)
}

func (l *Lifecycle) Shutdown(ctx context.Context) error {
	var errs []error

	for _, cleanup := range slices.Backward(l.cleanups) {
		errs = append(errs, cleanup(ctx))
	}

	return errors.Join(errs...)
}
