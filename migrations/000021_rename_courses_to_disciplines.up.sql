ALTER TABLE enrollments
DROP CONSTRAINT IF EXISTS enrollments_course_id_fkey;

ALTER TABLE invites
DROP CONSTRAINT IF EXISTS invites_course_id_fkey;

ALTER TABLE courses
RENAME TO disciplines;

ALTER TABLE enrollments
RENAME COLUMN course_id TO discipline_id;

ALTER TABLE invites
RENAME COLUMN course_id TO discipline_id;

ALTER TABLE disciplines
RENAME CONSTRAINT courses_pkey TO disciplines_pkey;

ALTER TABLE disciplines
RENAME CONSTRAINT courses_program_id_fkey TO disciplines_program_id_fkey;

ALTER INDEX IF EXISTS courses_pkey
RENAME TO disciplines_pkey;

ALTER TRIGGER trigger_update_timestamp_courses
ON disciplines
RENAME TO trigger_update_timestamp_disciplines;

ALTER TABLE enrollments
ADD CONSTRAINT enrollments_discipline_id_fkey
    FOREIGN KEY (discipline_id)
    REFERENCES disciplines(id)
    ON DELETE CASCADE;

ALTER TABLE invites
ADD CONSTRAINT invites_discipline_id_fkey
    FOREIGN KEY (discipline_id)
    REFERENCES disciplines(id)
    ON DELETE CASCADE;
