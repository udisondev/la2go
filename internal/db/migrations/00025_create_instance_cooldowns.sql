-- +goose Up

-- Per-character reentry cooldowns for instance zones.
-- Records when a player can next enter a given instance template.
CREATE TABLE IF NOT EXISTS instance_cooldowns (
    character_id  BIGINT NOT NULL,
    template_id   INTEGER NOT NULL,
    expire_time   BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (character_id, template_id),
    CONSTRAINT fk_instance_cooldown_char
        FOREIGN KEY (character_id) REFERENCES characters (character_id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_instance_cooldowns_expire ON instance_cooldowns (expire_time);

-- +goose Down

DROP TABLE IF EXISTS instance_cooldowns;
