-- +goose Up

-- Per-character Seven Signs participation data.
CREATE TABLE IF NOT EXISTS seven_signs (
    character_id   BIGINT PRIMARY KEY REFERENCES characters(character_id) ON DELETE CASCADE,
    cabal          VARCHAR(4)  NOT NULL DEFAULT '',
    seal           INTEGER     NOT NULL DEFAULT 0,
    red_stones     INTEGER     NOT NULL DEFAULT 0,
    green_stones   INTEGER     NOT NULL DEFAULT 0,
    blue_stones    INTEGER     NOT NULL DEFAULT 0,
    ancient_adena  BIGINT      NOT NULL DEFAULT 0,
    contribution   BIGINT      NOT NULL DEFAULT 0
);

-- Global Seven Signs status (single row, id=0).
CREATE TABLE IF NOT EXISTS seven_signs_status (
    id                  INTEGER     PRIMARY KEY DEFAULT 0,
    current_cycle       INTEGER     NOT NULL DEFAULT 1,
    festival_cycle      INTEGER     NOT NULL DEFAULT 0,
    active_period       INTEGER     NOT NULL DEFAULT 0,
    date_ms             BIGINT      NOT NULL DEFAULT 0,
    previous_winner     INTEGER     NOT NULL DEFAULT 0,
    dawn_stone_score    BIGINT      NOT NULL DEFAULT 0,
    dawn_festival_score INTEGER     NOT NULL DEFAULT 0,
    dusk_stone_score    BIGINT      NOT NULL DEFAULT 0,
    dusk_festival_score INTEGER     NOT NULL DEFAULT 0,
    avarice_owner       INTEGER     NOT NULL DEFAULT 0,
    gnosis_owner        INTEGER     NOT NULL DEFAULT 0,
    strife_owner        INTEGER     NOT NULL DEFAULT 0,
    avarice_dawn_score  INTEGER     NOT NULL DEFAULT 0,
    gnosis_dawn_score   INTEGER     NOT NULL DEFAULT 0,
    strife_dawn_score   INTEGER     NOT NULL DEFAULT 0,
    avarice_dusk_score  INTEGER     NOT NULL DEFAULT 0,
    gnosis_dusk_score   INTEGER     NOT NULL DEFAULT 0,
    strife_dusk_score   INTEGER     NOT NULL DEFAULT 0,
    accumulated_bonus0  INTEGER     NOT NULL DEFAULT 0,
    accumulated_bonus1  INTEGER     NOT NULL DEFAULT 0,
    accumulated_bonus2  INTEGER     NOT NULL DEFAULT 0,
    accumulated_bonus3  INTEGER     NOT NULL DEFAULT 0,
    accumulated_bonus4  INTEGER     NOT NULL DEFAULT 0
);

-- Insert default row if not exists.
INSERT INTO seven_signs_status (id) VALUES (0) ON CONFLICT DO NOTHING;

-- Festival results per tier per cabal per cycle.
CREATE TABLE IF NOT EXISTS seven_signs_festival (
    festival_id INTEGER     NOT NULL,
    cabal       VARCHAR(4)  NOT NULL,
    cycle       INTEGER     NOT NULL,
    date_ms     BIGINT      NOT NULL DEFAULT 0,
    score       INTEGER     NOT NULL DEFAULT 0,
    members     VARCHAR(255) NOT NULL DEFAULT '',
    PRIMARY KEY (festival_id, cabal, cycle)
);

-- Seed initial festival data for cycle 1.
INSERT INTO seven_signs_festival (festival_id, cabal, cycle) VALUES
    (0, 'dawn', 1), (1, 'dawn', 1), (2, 'dawn', 1), (3, 'dawn', 1), (4, 'dawn', 1),
    (0, 'dusk', 1), (1, 'dusk', 1), (2, 'dusk', 1), (3, 'dusk', 1), (4, 'dusk', 1)
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS seven_signs_festival;
DROP TABLE IF EXISTS seven_signs_status;
DROP TABLE IF EXISTS seven_signs;
