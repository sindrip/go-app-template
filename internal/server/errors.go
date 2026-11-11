package server

import (
	"errors"
	"net/http"

	"github.com/sindrip/go-app-template/internal/eventstore"
	"github.com/sindrip/go-app-template/internal/todo"
)

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, todo.ErrItemNotFound):
		status = http.StatusNotFound
	case errors.Is(err, todo.ErrItemAlreadyCompleted),
		errors.Is(err, eventstore.ErrVersionConflict):
		status = http.StatusConflict
	}

	writeStatus(w, status)
}

func writeStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
