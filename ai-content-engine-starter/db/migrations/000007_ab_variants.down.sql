DROP INDEX IF EXISTS drafts_source_channel_variant_uidx;
ALTER TABLE drafts DROP CONSTRAINT IF EXISTS drafts_variant_check;
ALTER TABLE drafts DROP COLUMN IF EXISTS variant;
