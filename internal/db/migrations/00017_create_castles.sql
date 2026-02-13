-- +goose Up

-- Castles (9 in Interlude)
CREATE TABLE IF NOT EXISTS castles (
    castle_id     INTEGER     PRIMARY KEY,
    name          VARCHAR(45) NOT NULL UNIQUE,
    owner_clan_id INTEGER     NOT NULL DEFAULT 0,
    tax_rate      SMALLINT    NOT NULL DEFAULT 0,
    treasury      BIGINT      NOT NULL DEFAULT 0,
    siege_date    TIMESTAMP   NULL,
    time_reg_over BOOLEAN     NOT NULL DEFAULT TRUE,
    show_npc_crest BOOLEAN    NOT NULL DEFAULT FALSE,
    ticket_buy_count INTEGER  NOT NULL DEFAULT 0
);

-- Siege clan registrations
CREATE TABLE IF NOT EXISTS siege_clans (
    castle_id     INTEGER  NOT NULL REFERENCES castles(castle_id) ON DELETE CASCADE,
    clan_id       INTEGER  NOT NULL,
    type          SMALLINT NOT NULL DEFAULT 0,
    PRIMARY KEY (clan_id, castle_id)
);

CREATE INDEX IF NOT EXISTS idx_siege_clans_castle ON siege_clans(castle_id);

-- Seed castles data
INSERT INTO castles (castle_id, name) VALUES
    (1, 'Gludio'),
    (2, 'Dion'),
    (3, 'Giran'),
    (4, 'Oren'),
    (5, 'Aden'),
    (6, 'Innadril'),
    (7, 'Goddard'),
    (8, 'Rune'),
    (9, 'Schuttgart')
ON CONFLICT (castle_id) DO NOTHING;

-- Add castle_id to clan_data
ALTER TABLE clan_data ADD COLUMN IF NOT EXISTS castle_id INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE clan_data DROP COLUMN IF EXISTS castle_id;
DROP TABLE IF EXISTS siege_clans;
DROP TABLE IF EXISTS castles;
