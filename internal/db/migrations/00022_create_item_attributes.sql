-- +goose Up
-- Phase 28: Augmentation System.
-- Stores weapon augmentation data per item instance.
-- Java reference: item_attributes table in L2J schema.
CREATE TABLE IF NOT EXISTS item_attributes (
    item_id         BIGINT PRIMARY KEY REFERENCES items(item_id) ON DELETE CASCADE,
    aug_attributes  INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_item_attributes_item_id ON item_attributes (item_id);

-- +goose Down
DROP TABLE IF EXISTS item_attributes;
