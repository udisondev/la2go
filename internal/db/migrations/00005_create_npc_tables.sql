-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS npc_templates (
    template_id INTEGER PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    title VARCHAR(100),
    level INTEGER NOT NULL DEFAULT 1 CHECK (level >= 1 AND level <= 100),
    max_hp INTEGER NOT NULL DEFAULT 1000 CHECK (max_hp > 0),
    max_mp INTEGER NOT NULL DEFAULT 500 CHECK (max_mp >= 0),
    p_atk INTEGER NOT NULL DEFAULT 0 CHECK (p_atk >= 0),
    p_def INTEGER NOT NULL DEFAULT 0 CHECK (p_def >= 0),
    m_atk INTEGER NOT NULL DEFAULT 0 CHECK (m_atk >= 0),
    m_def INTEGER NOT NULL DEFAULT 0 CHECK (m_def >= 0),
    aggro_range INTEGER NOT NULL DEFAULT 0 CHECK (aggro_range >= 0),
    move_speed INTEGER NOT NULL DEFAULT 80 CHECK (move_speed > 0),
    atk_speed INTEGER NOT NULL DEFAULT 253 CHECK (atk_speed > 0),
    respawn_min INTEGER NOT NULL DEFAULT 30 CHECK (respawn_min >= 0),
    respawn_max INTEGER NOT NULL DEFAULT 60 CHECK (respawn_max >= respawn_min)
);

CREATE TABLE IF NOT EXISTS spawns (
    spawn_id BIGSERIAL PRIMARY KEY,
    template_id INTEGER NOT NULL REFERENCES npc_templates(template_id) ON DELETE CASCADE,
    x INTEGER NOT NULL,
    y INTEGER NOT NULL,
    z INTEGER NOT NULL,
    heading SMALLINT NOT NULL DEFAULT 0 CHECK (heading >= 0 AND heading <= 65535),
    maximum_count INTEGER NOT NULL DEFAULT 1 CHECK (maximum_count >= 1),
    do_respawn BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_spawns_template_id ON spawns(template_id);
CREATE INDEX IF NOT EXISTS idx_spawns_location ON spawns(x, y, z);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_spawns_location;
DROP INDEX IF EXISTS idx_spawns_template_id;
DROP TABLE IF EXISTS spawns;
DROP TABLE IF EXISTS npc_templates;
-- +goose StatementEnd
