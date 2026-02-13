-- +goose Up
CREATE TABLE IF NOT EXISTS character_quests (
    character_id BIGINT NOT NULL REFERENCES characters(character_id) ON DELETE CASCADE,
    quest_name   VARCHAR(128) NOT NULL,
    variable     VARCHAR(64)  NOT NULL,
    value        VARCHAR(255) NOT NULL DEFAULT '',
    PRIMARY KEY (character_id, quest_name, variable)
);

CREATE INDEX idx_character_quests_char ON character_quests (character_id);
CREATE INDEX idx_character_quests_quest ON character_quests (quest_name);

-- +goose Down
DROP TABLE IF EXISTS character_quests;
