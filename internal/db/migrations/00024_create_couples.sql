-- +goose Up
CREATE TABLE IF NOT EXISTS couples (
    id SERIAL PRIMARY KEY,
    player1_id INT NOT NULL,
    player2_id INT NOT NULL,
    married BOOLEAN NOT NULL DEFAULT false,
    affianced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    married_at TIMESTAMPTZ,
    CONSTRAINT couples_player_order CHECK (player1_id < player2_id),
    CONSTRAINT couples_unique_pair UNIQUE (player1_id, player2_id)
);

CREATE INDEX idx_couples_player1 ON couples(player1_id);
CREATE INDEX idx_couples_player2 ON couples(player2_id);

-- +goose Down
DROP TABLE IF EXISTS couples;
