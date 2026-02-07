-- +goose Up
CREATE TABLE IF NOT EXISTS accounts (
    login        TEXT PRIMARY KEY,
    password     TEXT        NOT NULL,
    access_level INTEGER     NOT NULL DEFAULT 0,
    last_server  INTEGER     NOT NULL DEFAULT 1,
    last_ip      TEXT        NOT NULL DEFAULT '',
    last_active  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS accounts;
