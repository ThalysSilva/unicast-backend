DROP INDEX IF EXISTS idx_message_logs_delivery_group_id;
DROP INDEX IF EXISTS idx_message_logs_student_channel_created_at;

ALTER TABLE message_logs
DROP COLUMN IF EXISTS sender_address,
DROP COLUMN IF EXISTS sender_provider,
DROP COLUMN IF EXISTS sender_type,
DROP COLUMN IF EXISTS delivery_group_id;
