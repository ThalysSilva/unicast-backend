ALTER TABLE message_logs
DROP CONSTRAINT IF EXISTS message_logs_smtp_id_fkey;

ALTER TABLE message_logs
ADD CONSTRAINT message_logs_smtp_id_fkey
FOREIGN KEY (smtp_id) REFERENCES smtp_instances(id);

ALTER TABLE message_logs
DROP CONSTRAINT IF EXISTS message_logs_whatsapp_instance_id_fkey;

ALTER TABLE message_logs
ADD CONSTRAINT message_logs_whatsapp_instance_id_fkey
FOREIGN KEY (whatsapp_instance_id) REFERENCES whatsapp_instances(id);
