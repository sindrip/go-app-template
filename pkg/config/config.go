package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type Lookup func(string) (string, bool)

func (l Lookup) Load[T any]() (T, error) {
	return Loader[T]{lookup: l}.Load()
}

func (l Lookup) Defaults[T any](defaults T) Loader[T] {
	return Loader[T]{lookup: l, value: defaults}
}

type Loader[T any] struct {
	lookup Lookup
	value  T
}

func (l Loader[T]) Load() (T, error) {
	err := load(l.lookup, reflect.ValueOf(&l.value).Elem())

	return l.value, err
}

func load(lookup Lookup, v reflect.Value) error {
	var errs []error

	for i := range v.NumField() {
		field := v.Field(i)

		key, tagged := v.Type().Field(i).Tag.Lookup("env")
		if !tagged {
			if field.Kind() == reflect.Struct {
				errs = append(errs, load(lookup, field))
			}

			continue
		}

		text, ok := lookup(key)
		if !ok {
			continue
		}

		if err := set(field, text); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))
		}
	}

	return errors.Join(errs...)
}

func set(v reflect.Value, text string) error {
	if !v.CanSet() {
		return errors.New("field is unexported")
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(text)

	case reflect.Bool:
		b, err := strconv.ParseBool(text)
		if err != nil {
			return err
		}

		v.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(text, 10, v.Type().Bits())
		if err != nil {
			return err
		}

		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(text, 10, v.Type().Bits())
		if err != nil {
			return err
		}

		v.SetUint(n)

	default:
		return fmt.Errorf("unsupported type %s", v.Type())
	}

	return nil
}
