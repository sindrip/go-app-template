package server

import (
	"database/sql"
	"net/http"
	"uuid"

	"github.com/sindrip/go-app-template/internal/todo"
)

type listResponse struct {
	ID    uuid.UUID      `json:"id"`
	Items []itemResponse `json:"items"`
}

type itemResponse struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Done  bool      `json:"done"`
}

func newListResponse(id uuid.UUID, list todo.List) listResponse {
	items := make([]itemResponse, 0, len(list.Items))

	for _, item := range list.Items {
		items = append(items, itemResponse{ID: item.ID, Title: item.Title, Done: item.Done})
	}

	return listResponse{ID: id, Items: items}
}

func handleGetList(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listID, err := uuid.Parse(r.PathValue("listID"))
		if err != nil {
			writeStatus(w, http.StatusBadRequest)

			return
		}

		list, err := todo.Lists.Load(r.Context(), db, listID)
		if err != nil {
			writeError(w, err)

			return
		}

		encode(w, http.StatusOK, newListResponse(listID, list))
	})
}

func handleAddItem(db *sql.DB) http.Handler {
	type request struct {
		Title string `json:"title"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listID, err := uuid.Parse(r.PathValue("listID"))
		if err != nil {
			writeStatus(w, http.StatusBadRequest)

			return
		}

		req, err := decode[request](r)
		if err != nil {
			writeStatus(w, http.StatusBadRequest)

			return
		}

		itemID := uuid.New()

		list, err := todo.Lists.Decide(r.Context(), db, listID,
			func(list todo.List) ([]todo.Event, error) {
				return list.AddItem(itemID, req.Title)
			},
		)
		if err != nil {
			writeError(w, err)

			return
		}

		encode(w, http.StatusCreated, newListResponse(listID, list))
	})
}

func handleCompleteItem(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listID, err := uuid.Parse(r.PathValue("listID"))
		if err != nil {
			writeStatus(w, http.StatusBadRequest)

			return
		}

		itemID, err := uuid.Parse(r.PathValue("itemID"))
		if err != nil {
			writeStatus(w, http.StatusBadRequest)

			return
		}

		list, err := todo.Lists.Decide(r.Context(), db, listID,
			func(list todo.List) ([]todo.Event, error) {
				return list.CompleteItem(itemID)
			},
		)
		if err != nil {
			writeError(w, err)

			return
		}

		encode(w, http.StatusOK, newListResponse(listID, list))
	})
}
