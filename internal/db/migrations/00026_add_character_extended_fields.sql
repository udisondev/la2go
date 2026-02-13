-- +goose Up
-- Phase 34: Add missing character fields identified in audit report 11.
-- These fields are required for proper character persistence (title, base class,
-- noblesse, PvP stats, access level, appearance colors, recommendations, etc.)
-- Java reference: characters.sql (62 columns total)

-- Display & identity
ALTER TABLE characters ADD COLUMN IF NOT EXISTS title          VARCHAR(21)  NOT NULL DEFAULT '';
ALTER TABLE characters ADD COLUMN IF NOT EXISTS base_class     SMALLINT     NOT NULL DEFAULT 0;

-- PvP statistics
ALTER TABLE characters ADD COLUMN IF NOT EXISTS pvpkills       INTEGER      NOT NULL DEFAULT 0;

-- Status flags
ALTER TABLE characters ADD COLUMN IF NOT EXISTS nobless        BOOLEAN      NOT NULL DEFAULT false;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS hero           BOOLEAN      NOT NULL DEFAULT false;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS online         BOOLEAN      NOT NULL DEFAULT false;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS wants_peace    BOOLEAN      NOT NULL DEFAULT false;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS cancraft       BOOLEAN      NOT NULL DEFAULT true;

-- Access control
ALTER TABLE characters ADD COLUMN IF NOT EXISTS accesslevel    INTEGER      NOT NULL DEFAULT 0;

-- Appearance colors (L2 uses int32 RGB)
ALTER TABLE characters ADD COLUMN IF NOT EXISTS name_color     INTEGER      NOT NULL DEFAULT 16777215;  -- 0xFFFFFF = white
ALTER TABLE characters ADD COLUMN IF NOT EXISTS title_color    INTEGER      NOT NULL DEFAULT 15530402;  -- 0xECF9A2 = Java default

-- Recommendations
ALTER TABLE characters ADD COLUMN IF NOT EXISTS rec_have       SMALLINT     NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS rec_left       SMALLINT     NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS last_recom_date BIGINT      NOT NULL DEFAULT 0;

-- Experience & penalty
ALTER TABLE characters ADD COLUMN IF NOT EXISTS exp_before_death BIGINT     NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS death_penalty_level SMALLINT NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS fame           INTEGER      NOT NULL DEFAULT 0;

-- Clan extended
ALTER TABLE characters ADD COLUMN IF NOT EXISTS clan_privs     INTEGER      NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS subpledge      SMALLINT     NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN IF NOT EXISTS power_grade    SMALLINT     NOT NULL DEFAULT 0;

-- Online tracking
ALTER TABLE characters ADD COLUMN IF NOT EXISTS onlinetime     INTEGER      NOT NULL DEFAULT 0;

-- Teleport bookmarks
ALTER TABLE characters ADD COLUMN IF NOT EXISTS bookmarkslot   SMALLINT     NOT NULL DEFAULT 0;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_characters_online ON characters(online) WHERE online = true;

-- +goose Down
ALTER TABLE characters DROP COLUMN IF EXISTS bookmarkslot;
ALTER TABLE characters DROP COLUMN IF EXISTS onlinetime;
ALTER TABLE characters DROP COLUMN IF EXISTS power_grade;
ALTER TABLE characters DROP COLUMN IF EXISTS subpledge;
ALTER TABLE characters DROP COLUMN IF EXISTS clan_privs;
ALTER TABLE characters DROP COLUMN IF EXISTS fame;
ALTER TABLE characters DROP COLUMN IF EXISTS death_penalty_level;
ALTER TABLE characters DROP COLUMN IF EXISTS exp_before_death;
ALTER TABLE characters DROP COLUMN IF EXISTS last_recom_date;
ALTER TABLE characters DROP COLUMN IF EXISTS rec_left;
ALTER TABLE characters DROP COLUMN IF EXISTS rec_have;
ALTER TABLE characters DROP COLUMN IF EXISTS title_color;
ALTER TABLE characters DROP COLUMN IF EXISTS name_color;
ALTER TABLE characters DROP COLUMN IF EXISTS accesslevel;
ALTER TABLE characters DROP COLUMN IF EXISTS cancraft;
ALTER TABLE characters DROP COLUMN IF EXISTS wants_peace;
ALTER TABLE characters DROP COLUMN IF EXISTS online;
ALTER TABLE characters DROP COLUMN IF EXISTS hero;
ALTER TABLE characters DROP COLUMN IF EXISTS nobless;
ALTER TABLE characters DROP COLUMN IF EXISTS pvpkills;
ALTER TABLE characters DROP COLUMN IF EXISTS base_class;
ALTER TABLE characters DROP COLUMN IF EXISTS title;
