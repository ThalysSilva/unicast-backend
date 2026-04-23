ALTER TABLE message_logs
ADD COLUMN delivery_group_id UUID NULL,
ADD COLUMN sender_type VARCHAR NULL,
ADD COLUMN sender_provider VARCHAR NULL,
ADD COLUMN sender_address VARCHAR NULL;

CREATE INDEX idx_message_logs_student_channel_created_at
ON message_logs (student_id, channel, created_at DESC);

CREATE INDEX idx_message_logs_delivery_group_id
ON message_logs (delivery_group_id);
