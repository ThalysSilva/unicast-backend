CREATE TABLE programs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR NOT NULL,
    description VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    campus_id UUID NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (campus_id) REFERENCES campuses(id)
);