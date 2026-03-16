DROP INDEX IF EXISTS idx_performance_feedback_channel_variant_score;

ALTER TABLE performance_feedback DROP CONSTRAINT IF EXISTS performance_feedback_variant_check;
ALTER TABLE performance_feedback DROP COLUMN IF EXISTS variant;
