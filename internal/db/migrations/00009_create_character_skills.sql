-- +goose Up
CREATE TABLE character_skills (
    character_id BIGINT NOT NULL REFERENCES characters(character_id) ON DELETE CASCADE,
    skill_id     INTEGER NOT NULL,
    skill_level  INTEGER NOT NULL DEFAULT 1,
    class_index  INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (character_id, skill_id, class_index)
);
CREATE INDEX idx_character_skills_char ON character_skills(character_id);

-- +goose Down
DROP TABLE character_skills;
