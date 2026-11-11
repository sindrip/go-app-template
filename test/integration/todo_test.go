package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

type listJSON struct {
	ID    string     `json:"id"`
	Items []itemJSON `json:"items"`
}

type itemJSON struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func request(t *testing.T, method, url, body string) (int, listJSON) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), method, url, strings.NewReader(body))
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	var list listJSON

	if res.StatusCode < http.StatusBadRequest {
		require.NoError(t, json.Unmarshal(data, &list))
	}

	return res.StatusCode, list
}

func TestTodoList(t *testing.T) {
	t.Parallel()

	a := startApp(t, newDatabase(t))

	listID := uuid.New().String()
	listURL := serverURL(a, "/lists/"+listID)

	status, list := request(t, http.MethodPost, listURL+"/items", `{"title":"buy milk"}`)
	require.Equal(t, http.StatusCreated, status)
	require.Len(t, list.Items, 1)

	itemID := list.Items[0].ID
	completeURL := listURL + "/items/" + itemID + "/complete"

	require.Equal(t, listJSON{
		ID:    listID,
		Items: []itemJSON{{ID: itemID, Title: "buy milk", Done: false}},
	}, list)

	want := listJSON{
		ID:    listID,
		Items: []itemJSON{{ID: itemID, Title: "buy milk", Done: true}},
	}

	status, list = request(t, http.MethodPost, completeURL, "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, want, list)

	status, _ = request(t, http.MethodPost, completeURL, "")
	require.Equal(t, http.StatusConflict, status)

	unknownURL := listURL + "/items/" + uuid.New().String() + "/complete"

	status, _ = request(t, http.MethodPost, unknownURL, "")
	require.Equal(t, http.StatusNotFound, status)

	status, _ = request(t, http.MethodPost, listURL+"/items/not-a-uuid/complete", "")
	require.Equal(t, http.StatusBadRequest, status)

	status, _ = request(t, http.MethodPost, listURL+"/items", `not json`)
	require.Equal(t, http.StatusBadRequest, status)

	status, list = request(t, http.MethodGet, listURL, "")
	require.Equal(t, http.StatusOK, status)
	require.Equal(t, want, list)
}
