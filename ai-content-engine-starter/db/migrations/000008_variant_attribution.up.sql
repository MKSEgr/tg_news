ALTER TABLE performance_feedback ADD COLUMN IF NOT EXISTS variant TEXT;

UPDATE performance_feedback pf
SET variant = COALESCE(NULLIF(d.variant, ''), 'A')
FROM drafts d
WHERE d.id = pf.draft_id
  AND (pf.variant IS NULL OR pf.variant = '');

UPDATE performance_feedback
SET variant = 'A'
WHERE variant IS NULL OR variant = '';

ALTER TABLE performance_feedback ALTER COLUMN variant SET NOT NULL;
ALTER TABLE performance_feedback DROP CONSTRAINT IF EXISTS performance_feedback_variant_check;
ALTER TABLE performance_feedback ADD CONSTRAINT performance_feedback_variant_check CHECK (variant IN ('A', 'B'));

CREATE INDEX IF NOT EXISTS idx_performance_feedback_channel_variant_score
  ON performance_feedback(channel_id, variant, score DESC);
