-- 002_raw_ip_ua.sql
-- Store raw IP and User-Agent instead of hashes for better analytics

ALTER TABLE click_events RENAME COLUMN ip_hash TO ip;
ALTER TABLE click_events RENAME COLUMN ua_hash TO ua;
