-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS character_subclasses (
    character_id  BIGINT NOT NULL REFERENCES characters(character_id) ON DELETE CASCADE,
    class_id      INTEGER NOT NULL,
    class_index   SMALLINT NOT NULL CHECK (class_index >= 1 AND class_index <= 3),
    exp           BIGINT NOT NULL DEFAULT 0,
    sp            BIGINT NOT NULL DEFAULT 0,
    level         INTEGER NOT NULL DEFAULT 40,

    PRIMARY KEY (character_id, class_index),
    CONSTRAINT uq_character_subclass_class UNIQUE (character_id, class_id),
    CONSTRAINT chk_subclass_level CHECK (level >= 1 AND level <= 80)
);

CREATE INDEX idx_character_subclasses_character ON character_subclasses(character_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS character_subclasses;
-- +goose StatementEnd
