ALTER TABLE drafts ADD COLUMN IF NOT EXISTS variant TEXT NOT NULL DEFAULT 'A';

UPDATE drafts SET variant = 'A' WHERE variant IS NULL OR variant = '';

ALTER TABLE drafts DROP CONSTRAINT IF EXISTS drafts_variant_check;
ALTER TABLE drafts ADD CONSTRAINT drafts_variant_check CHECK (variant IN ('A', 'B'));

CREATE UNIQUE INDEX IF NOT EXISTS drafts_source_channel_variant_uidx
  ON drafts (source_item_id, channel_id, variant);
