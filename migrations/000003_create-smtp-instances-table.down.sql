CREATE TABLE smtp_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    host VARCHAR NOT NULL,
    port INTEGER NOT NULL,
    email VARCHAR NOT NULL,
    password VARCHAR NOT NULL,
    iv VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);