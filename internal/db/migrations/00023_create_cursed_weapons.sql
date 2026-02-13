-- +goose Up
-- Phase 32: Cursed Weapons System
-- Tracks active cursed weapons (Zariche, Akamanah) state between server restarts.
-- Java reference: cursed_weapons.sql
CREATE TABLE IF NOT EXISTS cursed_weapons (
    item_id         INTEGER PRIMARY KEY,                  -- 8190 (Zariche) or 8689 (Akamanah)
    char_id         INTEGER NOT NULL DEFAULT 0,           -- owner character object ID
    player_karma    INTEGER NOT NULL DEFAULT 0,           -- saved player karma before activation
    player_pk_kills INTEGER NOT NULL DEFAULT 0,           -- saved player PK count before activation
    nb_kills        INTEGER NOT NULL DEFAULT 0,           -- kill count with this weapon
    end_time        BIGINT  NOT NULL DEFAULT 0            -- Unix milliseconds expiration
);

CREATE INDEX IF NOT EXISTS idx_cursed_weapons_char ON cursed_weapons(char_id);

-- Phase 32: Add karma and pk_kills columns to characters table.
ALTER TABLE characters ADD COLUMN IF NOT EXISTS karma    INTEGER NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS pk_kills INTEGER NOT NULL DEFAULT 0;

-- +goose Down
DROP INDEX IF EXISTS idx_cursed_weapons_char;
DROP TABLE IF EXISTS cursed_weapons;
ALTER TABLE characters DROP COLUMN IF EXISTS karma;
ALTER TABLE characters DROP COLUMN IF EXISTS pk_kills;
