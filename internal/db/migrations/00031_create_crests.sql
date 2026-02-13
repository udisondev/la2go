-- +goose Up
CREATE TABLE IF NOT EXISTS crests (
    crest_id INT PRIMARY KEY,
    data BYTEA NOT NULL,
    type SMALLINT NOT NULL CHECK (type IN (1, 2, 3))
);

-- +goose Down
DROP TABLE IF EXISTS crests;
