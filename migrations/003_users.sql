-- 003_users.sql
-- User system for multi-tenant SaaS

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255),              -- NULL for OAuth users
    provider      VARCHAR(20) NOT NULL DEFAULT 'email', -- email/google/github
    provider_id   VARCHAR(255),              -- OAuth provider's user id
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    UNIQUE (provider, provider_id)
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_provider ON users (provider, provider_id) WHERE provider_id IS NOT NULL;
