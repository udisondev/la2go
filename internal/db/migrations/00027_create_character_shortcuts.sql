-- +goose Up
CREATE TABLE character_shortcuts (
    character_id BIGINT NOT NULL REFERENCES characters(character_id),
    slot SMALLINT NOT NULL CHECK (slot >= 0 AND slot <= 11),
    page SMALLINT NOT NULL CHECK (page >= 0 AND page <= 9),
    type SMALLINT NOT NULL CHECK (type >= 0 AND type <= 5),
    shortcut_id INT NOT NULL,
    level INT NOT NULL DEFAULT 0,
    class_index SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (character_id, slot, page, class_index)
);

-- +goose Down
DROP TABLE IF EXISTS character_shortcuts;
