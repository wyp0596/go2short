-- 002_api_tokens.sql
-- API tokens for external access

CREATE TABLE IF NOT EXISTS api_tokens (
    id           SERIAL PRIMARY KEY,
    token_hash   VARCHAR(64) NOT NULL UNIQUE,
    name         VARCHAR(100) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    disabled     BOOLEAN NOT NULL DEFAULT FALSE
);

-- Index for token lookup (only active tokens)
CREATE INDEX IF NOT EXISTS idx_api_tokens_hash ON api_tokens (token_hash) WHERE NOT disabled;
