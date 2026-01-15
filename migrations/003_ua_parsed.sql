-- 003_ua_parsed.sql
-- Add parsed UA fields for analytics

ALTER TABLE click_events ADD COLUMN device_type TEXT;
ALTER TABLE click_events ADD COLUMN browser TEXT;
ALTER TABLE click_events ADD COLUMN os TEXT;
