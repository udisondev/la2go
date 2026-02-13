-- +goose Up

-- Clan data
CREATE TABLE IF NOT EXISTS clan_data (
    clan_id       INTEGER PRIMARY KEY,
    clan_name     VARCHAR(45)  NOT NULL UNIQUE,
    leader_id     BIGINT       NOT NULL,
    clan_level    SMALLINT     NOT NULL DEFAULT 0,
    reputation    INTEGER      NOT NULL DEFAULT 0,
    crest_id      INTEGER      NOT NULL DEFAULT 0,
    large_crest_id INTEGER     NOT NULL DEFAULT 0,
    ally_id       INTEGER      NOT NULL DEFAULT 0,
    ally_name     VARCHAR(45)  NOT NULL DEFAULT '',
    ally_crest_id INTEGER      NOT NULL DEFAULT 0,
    notice        TEXT         NOT NULL DEFAULT '',
    notice_enabled BOOLEAN     NOT NULL DEFAULT FALSE,
    introduction  TEXT         NOT NULL DEFAULT '',
    dissolution_time BIGINT    NOT NULL DEFAULT 0,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clan_data_name ON clan_data(clan_name);
CREATE INDEX IF NOT EXISTS idx_clan_data_leader ON clan_data(leader_id);

-- Clan members
CREATE TABLE IF NOT EXISTS clan_members (
    character_id  BIGINT    PRIMARY KEY REFERENCES characters(character_id) ON DELETE CASCADE,
    clan_id       INTEGER   NOT NULL REFERENCES clan_data(clan_id) ON DELETE CASCADE,
    pledge_type   SMALLINT  NOT NULL DEFAULT 0,
    power_grade   SMALLINT  NOT NULL DEFAULT 5,
    title         VARCHAR(45) NOT NULL DEFAULT '',
    sponsor_id    BIGINT    NOT NULL DEFAULT 0,
    apprentice    BIGINT    NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_clan_members_clan ON clan_members(clan_id);

-- Clan sub-pledges
CREATE TABLE IF NOT EXISTS clan_subpledges (
    clan_id       INTEGER   NOT NULL REFERENCES clan_data(clan_id) ON DELETE CASCADE,
    sub_pledge_id SMALLINT  NOT NULL,
    name          VARCHAR(45) NOT NULL,
    leader_id     BIGINT    NOT NULL DEFAULT 0,
    PRIMARY KEY (clan_id, sub_pledge_id)
);

-- Clan wars
CREATE TABLE IF NOT EXISTS clan_wars (
    clan1_id      INTEGER   NOT NULL REFERENCES clan_data(clan_id) ON DELETE CASCADE,
    clan2_id      INTEGER   NOT NULL REFERENCES clan_data(clan_id) ON DELETE CASCADE,
    started_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (clan1_id, clan2_id)
);

-- Clan skills
CREATE TABLE IF NOT EXISTS clan_skills (
    clan_id       INTEGER   NOT NULL REFERENCES clan_data(clan_id) ON DELETE CASCADE,
    skill_id      INTEGER   NOT NULL,
    skill_level   SMALLINT  NOT NULL DEFAULT 1,
    PRIMARY KEY (clan_id, skill_id)
);

-- Clan rank privileges
CREATE TABLE IF NOT EXISTS clan_privs (
    clan_id       INTEGER   NOT NULL REFERENCES clan_data(clan_id) ON DELETE CASCADE,
    power_grade   SMALLINT  NOT NULL,
    privs         INTEGER   NOT NULL DEFAULT 0,
    PRIMARY KEY (clan_id, power_grade)
);

-- Add clan_id to characters table
ALTER TABLE characters ADD COLUMN IF NOT EXISTS clan_id INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE characters DROP COLUMN IF EXISTS clan_id;
DROP TABLE IF EXISTS clan_privs;
DROP TABLE IF EXISTS clan_skills;
DROP TABLE IF EXISTS clan_wars;
DROP TABLE IF EXISTS clan_subpledges;
DROP TABLE IF EXISTS clan_members;
DROP TABLE IF EXISTS clan_data;
