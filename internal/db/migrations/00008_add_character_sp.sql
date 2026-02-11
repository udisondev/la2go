-- +goose Up
ALTER TABLE characters ADD COLUMN sp BIGINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE characters DROP COLUMN sp;
