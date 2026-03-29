ALTER TABLE students
DROP CONSTRAINT IF EXISTS students_status_check;

ALTER TABLE students
ADD CONSTRAINT students_status_check
CHECK (status IN ('ACTIVE', 'CANCELED', 'GRADUATED', 'LOCKED', 'PENDING'));
