-- +goose Up
CREATE TABLE character_friends (
    char_id    BIGINT   NOT NULL REFERENCES characters(character_id),
    friend_id  BIGINT   NOT NULL,
    relation   SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (char_id, friend_id),
    CHECK (relation IN (0, 1))
);

CREATE INDEX idx_character_friends_lookup ON character_friends(char_id, relation);

-- +goose Down
DROP TABLE IF EXISTS character_friends;
