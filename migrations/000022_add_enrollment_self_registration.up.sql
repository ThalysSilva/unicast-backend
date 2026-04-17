ALTER TABLE enrollments
ADD COLUMN self_registration_completed_at TIMESTAMPTZ NULL,
ADD COLUMN self_registration_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE enrollments
ADD CONSTRAINT enrollments_discipline_student_unique
    UNIQUE (discipline_id, student_id);
