-- +goose Up
-- Phase 23: Raid Boss System tables

-- Raid boss spawn tracking — stores respawn times for raid bosses
CREATE TABLE IF NOT EXISTS raidboss_spawnlist (
    boss_id         INTEGER PRIMARY KEY,          -- NPC template ID
    respawn_time    BIGINT  NOT NULL DEFAULT 0,   -- Unix timestamp (seconds) of next respawn
    current_hp      DOUBLE PRECISION NOT NULL DEFAULT 0,
    current_mp      DOUBLE PRECISION NOT NULL DEFAULT 0
);

-- Grand boss state tracking — stores status and respawn data
CREATE TABLE IF NOT EXISTS grandboss_data (
    boss_id         INTEGER PRIMARY KEY,          -- NPC template ID (unique grand boss)
    loc_x           INTEGER NOT NULL DEFAULT 0,
    loc_y           INTEGER NOT NULL DEFAULT 0,
    loc_z           INTEGER NOT NULL DEFAULT 0,
    heading         INTEGER NOT NULL DEFAULT 0,
    current_hp      DOUBLE PRECISION NOT NULL DEFAULT 0,
    current_mp      DOUBLE PRECISION NOT NULL DEFAULT 0,
    respawn_time    BIGINT  NOT NULL DEFAULT 0,   -- Unix timestamp (seconds)
    status          SMALLINT NOT NULL DEFAULT 0   -- 0=ALIVE, 1=DEAD, 2=FIGHTING, 3=WAITING
);

-- Raid boss kill points per character
CREATE TABLE IF NOT EXISTS character_raid_points (
    character_id    INTEGER NOT NULL REFERENCES characters(character_id) ON DELETE CASCADE,
    boss_id         INTEGER NOT NULL,             -- NPC template ID of killed raid boss
    points          INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (character_id, boss_id)
);

CREATE INDEX IF NOT EXISTS idx_raid_points_character ON character_raid_points(character_id);

-- +goose Down
DROP TABLE IF EXISTS character_raid_points;
DROP TABLE IF EXISTS grandboss_data;
DROP TABLE IF EXISTS raidboss_spawnlist;
