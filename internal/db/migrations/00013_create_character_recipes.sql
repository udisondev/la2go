-- +goose Up
-- Phase 15: Recipe/Craft System.
-- Stores learned recipes per character.

CREATE TABLE IF NOT EXISTS character_recipes (
    character_id BIGINT NOT NULL REFERENCES characters(character_id) ON DELETE CASCADE,
    recipe_id    INTEGER NOT NULL,
    type         SMALLINT NOT NULL DEFAULT 0, -- 0=common, 1=dwarven
    PRIMARY KEY (character_id, recipe_id)
);

CREATE INDEX IF NOT EXISTS idx_character_recipes_char ON character_recipes(character_id);

-- +goose Down
DROP TABLE IF EXISTS character_recipes;
