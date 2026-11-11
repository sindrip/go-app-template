package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func encode[T any](w http.ResponseWriter, status int, value T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func decode[T any](r *http.Request) (T, error) {
	var value T

	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		return value, fmt.Errorf("decode request: %w", err)
	}

	return value, nil
}
