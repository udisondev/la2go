-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS characters (
    character_id  BIGSERIAL PRIMARY KEY,
    account_id    BIGINT NOT NULL,
    name          VARCHAR(50) NOT NULL UNIQUE,
    level         INTEGER NOT NULL DEFAULT 1,
    race_id       INTEGER NOT NULL,
    class_id      INTEGER NOT NULL,

    -- Location
    x             INTEGER NOT NULL DEFAULT 0,
    y             INTEGER NOT NULL DEFAULT 0,
    z             INTEGER NOT NULL DEFAULT 0,
    heading       SMALLINT NOT NULL DEFAULT 0 CHECK (heading >= 0 AND heading <= 65535),

    -- Stats
    current_hp    INTEGER NOT NULL,
    max_hp        INTEGER NOT NULL,
    current_mp    INTEGER NOT NULL,
    max_mp        INTEGER NOT NULL,
    current_cp    INTEGER NOT NULL,
    max_cp        INTEGER NOT NULL,

    experience    BIGINT NOT NULL DEFAULT 0,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login    TIMESTAMPTZ,

    CONSTRAINT chk_level CHECK (level >= 1 AND level <= 80),
    CONSTRAINT chk_name_length CHECK (LENGTH(name) >= 2)
    -- TODO: Add proper FK constraint when accounts table is refactored to use account_id BIGINT
);

CREATE INDEX idx_characters_account_id ON characters(account_id);
CREATE INDEX idx_characters_name ON characters(LOWER(name));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS characters;
-- +goose StatementEnd
