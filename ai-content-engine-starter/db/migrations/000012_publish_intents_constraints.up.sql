ALTER TABLE publish_intents
ADD CONSTRAINT publish_intents_status_check
CHECK (status IN ('planned', 'skipped', 'cancelled'));

CREATE UNIQUE INDEX IF NOT EXISTS ux_publish_intents_raw_item_channel
ON publish_intents (raw_item_id, channel_id);
