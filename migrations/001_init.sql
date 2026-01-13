-- 001_init.sql
-- Initial schema for go2short

-- Links table: source of truth for short URLs
CREATE TABLE IF NOT EXISTS links (
    code        TEXT PRIMARY KEY,
    long_url    TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ,
    is_disabled BOOLEAN NOT NULL DEFAULT FALSE
);

-- Index for expiration queries
CREATE INDEX IF NOT EXISTS idx_links_expires_at ON links (expires_at) WHERE expires_at IS NOT NULL;

-- Click events table
CREATE TABLE IF NOT EXISTS click_events (
    id          BIGSERIAL PRIMARY KEY,
    code        TEXT NOT NULL,
    ts          TIMESTAMPTZ NOT NULL,
    ip_hash     TEXT,
    ua_hash     TEXT,
    referer     TEXT
);

-- Index for aggregation queries (code + timestamp)
CREATE INDEX IF NOT EXISTS idx_clicks_code_ts ON click_events (code, ts);

-- Index for time range queries (trend, top links)
CREATE INDEX IF NOT EXISTS idx_clicks_ts ON click_events (ts);
