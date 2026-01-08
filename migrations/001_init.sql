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

-- Click events table: partitioned by time for efficient data management
CREATE TABLE IF NOT EXISTS click_events (
    id          BIGSERIAL,
    code        TEXT NOT NULL,
    ts          TIMESTAMPTZ NOT NULL,
    ip_hash     TEXT,
    ua_hash     TEXT,
    referer     TEXT,
    PRIMARY KEY (id, ts)
) PARTITION BY RANGE (ts);

-- Create initial partition for current month
CREATE TABLE IF NOT EXISTS click_events_default PARTITION OF click_events DEFAULT;

-- Index for aggregation queries (code + timestamp)
CREATE INDEX IF NOT EXISTS idx_clicks_code_ts ON click_events (code, ts);

-- Index for time range queries (trend, top links)
CREATE INDEX IF NOT EXISTS idx_clicks_ts ON click_events (ts);

-- Helper function to create monthly partitions
CREATE OR REPLACE FUNCTION create_click_partition(partition_date DATE)
RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    start_date := DATE_TRUNC('month', partition_date);
    end_date := start_date + INTERVAL '1 month';
    partition_name := 'click_events_' || TO_CHAR(start_date, 'YYYY_MM');

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF click_events
         FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );
END;
$$ LANGUAGE plpgsql;

-- Create partitions for next 3 months
SELECT create_click_partition(CURRENT_DATE);
SELECT create_click_partition((CURRENT_DATE + INTERVAL '1 month')::DATE);
SELECT create_click_partition((CURRENT_DATE + INTERVAL '2 months')::DATE);
