ALTER TABLE topic_memory
    DROP CONSTRAINT IF EXISTS topic_memory_topic_not_empty,
    DROP CONSTRAINT IF EXISTS topic_memory_mention_count_positive;
