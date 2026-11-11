package server

import (
	"database/sql"
	"net/http"
)

func addRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.Handle("GET /lists/{listID}", handleGetList(db))
	mux.Handle("POST /lists/{listID}/items", handleAddItem(db))
	mux.Handle("POST /lists/{listID}/items/{itemID}/complete", handleCompleteItem(db))
}
