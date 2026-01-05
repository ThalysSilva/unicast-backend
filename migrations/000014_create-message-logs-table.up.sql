CREATE TABLE message_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id),
    channel VARCHAR NOT NULL, -- EMAIL ou WHATSAPP
    success BOOLEAN NOT NULL,
    error_text TEXT,
    subject TEXT,
    body TEXT,
    smtp_id UUID NULL REFERENCES smtp_instances(id),
    whatsapp_instance_id UUID NULL REFERENCES whatsapp_instances(id),
    attachment_names TEXT,
    attachment_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
