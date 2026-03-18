CREATE UNIQUE INDEX ux_content_rules_global_kind_pattern
    ON content_rules(kind, pattern)
    WHERE channel_id IS NULL;
