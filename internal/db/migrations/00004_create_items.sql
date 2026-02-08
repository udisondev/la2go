-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS items (
    item_id       BIGSERIAL PRIMARY KEY,
    owner_id      BIGINT NOT NULL,
    item_type     INTEGER NOT NULL,
    count         INTEGER NOT NULL DEFAULT 1 CHECK (count > 0),
    enchant       INTEGER NOT NULL DEFAULT 0 CHECK (enchant >= 0),
    location      INTEGER NOT NULL DEFAULT 0,
    slot_id       INTEGER NOT NULL DEFAULT -1,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_owner FOREIGN KEY (owner_id) REFERENCES characters(character_id) ON DELETE CASCADE
);

CREATE INDEX idx_items_owner_id ON items(owner_id);
CREATE INDEX idx_items_location ON items(owner_id, location);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
-- +goose StatementEnd
