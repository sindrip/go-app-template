package server

import (
	"database/sql"
	"net/http"
)

func New(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, db)

	return mux
}
