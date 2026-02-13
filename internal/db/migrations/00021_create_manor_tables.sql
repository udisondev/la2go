-- +goose Up
CREATE TABLE IF NOT EXISTS castle_manor_production (
    castle_id   INT     NOT NULL,
    seed_id     INT     NOT NULL,
    amount      INT     NOT NULL DEFAULT 0,
    start_amount INT    NOT NULL DEFAULT 0,
    price       BIGINT  NOT NULL DEFAULT 0,
    next_period BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (castle_id, seed_id, next_period)
);

CREATE TABLE IF NOT EXISTS castle_manor_procure (
    castle_id   INT     NOT NULL,
    crop_id     INT     NOT NULL,
    amount      INT     NOT NULL DEFAULT 0,
    start_amount INT    NOT NULL DEFAULT 0,
    price       BIGINT  NOT NULL DEFAULT 0,
    reward_type INT     NOT NULL DEFAULT 0,
    next_period BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (castle_id, crop_id, next_period)
);

-- +goose Down
DROP TABLE IF EXISTS castle_manor_procure;
DROP TABLE IF EXISTS castle_manor_production;
