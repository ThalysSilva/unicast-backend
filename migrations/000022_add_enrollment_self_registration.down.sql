ALTER TABLE enrollments
DROP CONSTRAINT IF EXISTS enrollments_discipline_student_unique;

ALTER TABLE enrollments
DROP COLUMN IF EXISTS self_registration_count,
DROP COLUMN IF EXISTS self_registration_completed_at;
