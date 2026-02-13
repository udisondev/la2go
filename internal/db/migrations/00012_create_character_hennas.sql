-- +goose Up
-- Phase 13: Henna System — tattoo slots per character.
-- PK: (character_id, slot, class_index) — one henna per slot per subclass.

CREATE TABLE IF NOT EXISTS character_hennas (
    character_id BIGINT NOT NULL REFERENCES characters(character_id) ON DELETE CASCADE,
    slot         SMALLINT NOT NULL CHECK (slot BETWEEN 1 AND 3),
    dye_id       INTEGER NOT NULL,
    class_index  SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (character_id, slot, class_index)
);

CREATE INDEX IF NOT EXISTS idx_character_hennas_char_class
    ON character_hennas (character_id, class_index);

-- +goose Down
DROP TABLE IF EXISTS character_hennas;
