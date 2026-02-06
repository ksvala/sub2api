-- Track platform/model/version snapshot used for billing audit and replay.
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS provider VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS billing_model VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS price_version VARCHAR(128) NOT NULL DEFAULT '';

-- Backfill historical rows for compatibility.
UPDATE usage_logs
SET billing_model = model
WHERE billing_model = '';

UPDATE usage_logs
SET provider = 'unknown'
WHERE provider = '';

UPDATE usage_logs
SET price_version = 'pricing:legacy'
WHERE price_version = '';
