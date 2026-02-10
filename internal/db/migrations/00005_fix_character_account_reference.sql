-- +goose Up
-- +goose StatementBegin
-- Fix account reference in characters table
-- Change from account_id BIGINT to account_name TEXT to match accounts.login

ALTER TABLE characters
  DROP COLUMN account_id,
  ADD COLUMN account_name TEXT NOT NULL DEFAULT '';

-- Add FK constraint to accounts table
ALTER TABLE characters
  ADD CONSTRAINT fk_characters_account
  FOREIGN KEY (account_name) REFERENCES accounts(login)
  ON DELETE CASCADE;

-- Update index
DROP INDEX IF EXISTS idx_characters_account_id;
CREATE INDEX idx_characters_account_name ON characters(account_name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE characters
  DROP CONSTRAINT IF EXISTS fk_characters_account,
  DROP COLUMN account_name,
  ADD COLUMN account_id BIGINT NOT NULL DEFAULT 0;

DROP INDEX IF EXISTS idx_characters_account_name;
CREATE INDEX idx_characters_account_id ON characters(account_id);
-- +goose StatementEnd
