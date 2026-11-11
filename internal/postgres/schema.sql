CREATE TABLE events (
    stream_id  uuid        NOT NULL,
    version    integer     NOT NULL,
    type       text        NOT NULL,
    data       jsonb       NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (stream_id, version)
);
