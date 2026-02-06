-- Persist billing price source for auditability (model_pricing/group_image_price/fallback).
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS price_source VARCHAR(64) NOT NULL DEFAULT '';

-- Backfill historical rows to a deterministic value.
UPDATE usage_logs
SET price_source = 'legacy'
WHERE price_source = '';
