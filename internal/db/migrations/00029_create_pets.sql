-- +goose Up
CREATE TABLE IF NOT EXISTS pets (
    item_obj_id   INTEGER PRIMARY KEY,
    name          VARCHAR(16),
    level         SMALLINT NOT NULL DEFAULT 1,
    cur_hp        INTEGER NOT NULL DEFAULT 0,
    cur_mp        INTEGER NOT NULL DEFAULT 0,
    exp           BIGINT NOT NULL DEFAULT 0,
    sp            INTEGER NOT NULL DEFAULT 0,
    fed           INTEGER NOT NULL DEFAULT 0,
    owner_id      INTEGER NOT NULL DEFAULT 0,
    restore       BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_pets_owner_id ON pets (owner_id);

-- +goose Down
DROP TABLE IF EXISTS pets;
