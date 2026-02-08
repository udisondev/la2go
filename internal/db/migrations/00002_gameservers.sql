-- +goose Up
CREATE TABLE IF NOT EXISTS gameservers (
    server_id INTEGER PRIMARY KEY,
    hexid TEXT NOT NULL UNIQUE,
    host TEXT
);

-- Вставляем тестовый сервер
INSERT INTO gameservers (server_id, hexid, host) VALUES
    (1, '00000000000000000000000000000001', '127.0.0.1');

-- +goose Down
DROP TABLE IF EXISTS gameservers;
