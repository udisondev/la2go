-- +goose Up
CREATE TABLE IF NOT EXISTS character_summons (
    owner_id        INTEGER NOT NULL,
    summon_skill_id INTEGER NOT NULL,
    cur_hp          INTEGER NOT NULL DEFAULT 0,
    cur_mp          INTEGER NOT NULL DEFAULT 0,
    remaining_time  INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (owner_id, summon_skill_id)
);

-- +goose Down
DROP TABLE IF EXISTS character_summons;
