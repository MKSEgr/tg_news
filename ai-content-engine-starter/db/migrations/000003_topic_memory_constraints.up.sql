ALTER TABLE topic_memory
    ADD CONSTRAINT topic_memory_topic_not_empty CHECK (BTRIM(topic) <> ''),
    ADD CONSTRAINT topic_memory_mention_count_positive CHECK (mention_count > 0);
