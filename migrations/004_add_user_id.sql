-- 004_add_user_id.sql
-- Add user_id column to links and api_tokens for multi-tenant isolation

ALTER TABLE links ADD COLUMN IF NOT EXISTS user_id INT REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_links_user_id ON links (user_id);

ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS user_id INT REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_api_tokens_user_id ON api_tokens (user_id);
