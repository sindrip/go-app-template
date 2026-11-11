package lifecycle

import (
	"context"
	"fmt"
)

type PingCloser interface {
	PingContext(ctx context.Context) error
	Close() error
}

func (l *Lifecycle) Database[T PingCloser](db T, err error) (T, error) {
	if err != nil {
		return db, err
	}

	l.Check(fmt.Sprintf("%T", db), db.PingContext)

	l.cleanups = append(l.cleanups, func(context.Context) error {
		return db.Close()
	})

	return db, nil
}
