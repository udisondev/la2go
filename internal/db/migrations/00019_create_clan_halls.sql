-- +goose Up

-- Clan halls (auctionable).
CREATE TABLE IF NOT EXISTS clan_halls (
    hall_id       INTEGER PRIMARY KEY,
    name          VARCHAR(100) NOT NULL DEFAULT '',
    owner_clan_id INTEGER NOT NULL DEFAULT 0,
    lease         INTEGER NOT NULL DEFAULT 0,
    paid_until    BIGINT NOT NULL DEFAULT 0,
    grade         SMALLINT NOT NULL DEFAULT 0,
    location      VARCHAR(50) NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_clan_halls_owner ON clan_halls (owner_clan_id);

-- Clan hall functions (upgrades purchased by clan).
CREATE TABLE IF NOT EXISTS clan_hall_functions (
    hall_id   INTEGER NOT NULL,
    func_type SMALLINT NOT NULL,
    level     SMALLINT NOT NULL DEFAULT 0,
    lease     INTEGER NOT NULL DEFAULT 0,
    rate      BIGINT NOT NULL DEFAULT 0,
    end_time  BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (hall_id, func_type)
);

-- Auctions for free auctionable halls.
CREATE TABLE IF NOT EXISTS clan_hall_auctions (
    auction_id   INTEGER PRIMARY KEY,
    hall_id      INTEGER NOT NULL,
    starting_bid BIGINT NOT NULL DEFAULT 0,
    end_date     BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ch_auctions_hall ON clan_hall_auctions (hall_id);

-- Auction bids.
CREATE TABLE IF NOT EXISTS clan_hall_auction_bids (
    auction_id INTEGER NOT NULL,
    clan_id    INTEGER NOT NULL,
    max_bid    BIGINT NOT NULL DEFAULT 0,
    cur_bid    BIGINT NOT NULL DEFAULT 0,
    bid_time   BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (auction_id, clan_id)
);

-- Link clan to clan hall (add column to existing clan_data table).
ALTER TABLE clan_data ADD COLUMN IF NOT EXISTS clan_hall_id INTEGER NOT NULL DEFAULT 0;

-- +goose Down

ALTER TABLE clan_data DROP COLUMN IF EXISTS clan_hall_id;

DROP TABLE IF EXISTS clan_hall_auction_bids;
DROP TABLE IF EXISTS clan_hall_auctions;
DROP TABLE IF EXISTS clan_hall_functions;
DROP TABLE IF EXISTS clan_halls;
