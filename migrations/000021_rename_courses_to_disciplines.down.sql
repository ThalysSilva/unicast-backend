ALTER TABLE enrollments
DROP CONSTRAINT IF EXISTS enrollments_discipline_id_fkey;

ALTER TABLE invites
DROP CONSTRAINT IF EXISTS invites_discipline_id_fkey;

ALTER TABLE disciplines
RENAME TO courses;

ALTER TABLE enrollments
RENAME COLUMN discipline_id TO course_id;

ALTER TABLE invites
RENAME COLUMN discipline_id TO course_id;

ALTER TABLE courses
RENAME CONSTRAINT disciplines_pkey TO courses_pkey;

ALTER TABLE courses
RENAME CONSTRAINT disciplines_program_id_fkey TO courses_program_id_fkey;

ALTER INDEX IF EXISTS disciplines_pkey
RENAME TO courses_pkey;

ALTER TRIGGER trigger_update_timestamp_disciplines
ON courses
RENAME TO trigger_update_timestamp_courses;

ALTER TABLE enrollments
ADD CONSTRAINT enrollments_course_id_fkey
    FOREIGN KEY (course_id)
    REFERENCES courses(id)
    ON DELETE CASCADE;

ALTER TABLE invites
ADD CONSTRAINT invites_course_id_fkey
    FOREIGN KEY (course_id)
    REFERENCES courses(id)
    ON DELETE CASCADE;
