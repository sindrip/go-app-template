-- name: AppendEvent :exec
INSERT INTO events (stream_id, version, type, data)
VALUES ($1, $2, $3, $4);

-- name: ListEvents :many
SELECT stream_id, version, type, data, created_at
FROM events
WHERE stream_id = $1
ORDER BY version;
