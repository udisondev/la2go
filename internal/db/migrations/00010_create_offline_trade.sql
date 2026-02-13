-- +goose Up
-- Phase 8.1: Private Store System — offline trade tables.

CREATE TABLE IF NOT EXISTS character_offline_trade (
    char_id      BIGINT NOT NULL PRIMARY KEY REFERENCES characters(character_id) ON DELETE CASCADE,
    created_at   BIGINT NOT NULL DEFAULT 0,
    store_type   SMALLINT NOT NULL DEFAULT 0,
    title        VARCHAR(50) DEFAULT ''
);

CREATE TABLE IF NOT EXISTS character_offline_trade_items (
    char_id      BIGINT NOT NULL REFERENCES characters(character_id) ON DELETE CASCADE,
    item         INTEGER NOT NULL DEFAULT 0,
    count        BIGINT NOT NULL DEFAULT 0,
    price        BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_offline_trade_items_char ON character_offline_trade_items(char_id);

-- +goose Down
DROP TABLE IF EXISTS character_offline_trade_items;
DROP TABLE IF EXISTS character_offline_trade;
