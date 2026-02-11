-- +goose Up
ALTER TABLE npc_templates ADD COLUMN base_exp BIGINT NOT NULL DEFAULT 0;
ALTER TABLE npc_templates ADD COLUMN base_sp INTEGER NOT NULL DEFAULT 0;
UPDATE npc_templates SET base_exp = level * level * 10, base_sp = level * 5;

-- +goose Down
ALTER TABLE npc_templates DROP COLUMN base_exp;
ALTER TABLE npc_templates DROP COLUMN base_sp;
