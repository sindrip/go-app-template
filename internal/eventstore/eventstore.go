package eventstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"uuid"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sindrip/go-app-template/internal/postgres"
)

const uniqueViolation = "23505"

var ErrVersionConflict = errors.New("eventstore: version conflict")

type Stream[S any, E any] struct {
	apply    func(S, E) S
	names    map[reflect.Type]string
	decoders map[string]func([]byte) (E, error)
}

func NewStream[S any, E any](apply func(S, E) S) *Stream[S, E] {
	return &Stream[S, E]{
		apply:    apply,
		names:    make(map[reflect.Type]string),
		decoders: make(map[string]func([]byte) (E, error)),
	}
}

func (s *Stream[S, E]) Event[T any](name string) *Stream[S, E] {
	var zero T

	if _, ok := any(zero).(E); !ok {
		panic(fmt.Sprintf("eventstore: %T is not an event of this stream", zero))
	}

	s.names[reflect.TypeFor[T]()] = name
	s.decoders[name] = func(payload []byte) (E, error) {
		var event T

		err := json.Unmarshal(payload, &event)

		return any(event).(E), err
	}

	return s
}

func (s *Stream[S, E]) Load(ctx context.Context, db *sql.DB, id uuid.UUID) (S, error) {
	state, _, err := s.load(ctx, postgres.New(db), id)

	return state, err
}

func (s *Stream[S, E]) Decide(
	ctx context.Context,
	db *sql.DB,
	id uuid.UUID,
	decide func(S) ([]E, error),
) (S, error) {
	var zero S

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}

	defer func() { _ = tx.Rollback() }()

	queries := postgres.New(tx)

	state, version, err := s.load(ctx, queries, id)
	if err != nil {
		return zero, err
	}

	events, err := decide(state)
	if err != nil {
		return zero, err
	}

	for _, event := range events {
		version++

		if err := s.append(ctx, queries, id, version, event); err != nil {
			return zero, err
		}

		state = s.apply(state, event)
	}

	if err := tx.Commit(); err != nil {
		return zero, err
	}

	return state, nil
}

func (s *Stream[S, E]) load(
	ctx context.Context,
	queries *postgres.Queries,
	id uuid.UUID,
) (S, int32, error) {
	var (
		state   S
		version int32
	)

	rows, err := queries.ListEvents(ctx, id)
	if err != nil {
		return state, 0, err
	}

	for _, row := range rows {
		decode, ok := s.decoders[row.Type]
		if !ok {
			return state, 0, fmt.Errorf("eventstore: unknown event type %q", row.Type)
		}

		event, err := decode(row.Data)
		if err != nil {
			return state, 0, fmt.Errorf("eventstore: decode %s: %w", row.Type, err)
		}

		state = s.apply(state, event)
		version = row.Version
	}

	return state, version, nil
}

func (s *Stream[S, E]) append(
	ctx context.Context,
	queries *postgres.Queries,
	id uuid.UUID,
	version int32,
	event E,
) error {
	name, ok := s.names[reflect.TypeOf(event)]
	if !ok {
		return fmt.Errorf("eventstore: unregistered event %T", event)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("eventstore: encode %s: %w", name, err)
	}

	err = queries.AppendEvent(ctx, postgres.AppendEventParams{
		StreamID: id,
		Version:  version,
		Type:     name,
		Data:     data,
	})
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
		return ErrVersionConflict
	}

	return err
}
