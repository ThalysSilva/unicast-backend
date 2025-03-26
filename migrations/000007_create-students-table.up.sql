CREATE TABLE students (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id VARCHAR UNIQUE NOT NULL,
    name VARCHAR,
    phone VARCHAR,
    email VARCHAR,
    annotation VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR NOT NULL CHECK (status IN ('ACTIVE', 'CANCELED', 'GRADUATED', 'LOCKED'))
);