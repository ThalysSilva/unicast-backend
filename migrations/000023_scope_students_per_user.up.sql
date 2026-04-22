ALTER TABLE students
ADD COLUMN user_owner_id UUID NULL;

CREATE TEMP TABLE tmp_student_owner_map AS
SELECT DISTINCT e.student_id AS original_student_id, ca.user_owner_id
FROM enrollments e
JOIN disciplines d ON d.id = e.discipline_id
JOIN programs p ON p.id = d.program_id
JOIN campuses ca ON ca.id = p.campus_id;

WITH primary_owner AS (
    SELECT DISTINCT ON (original_student_id) original_student_id, user_owner_id
    FROM tmp_student_owner_map
    ORDER BY original_student_id, user_owner_id
)
UPDATE students s
SET user_owner_id = po.user_owner_id
FROM primary_owner po
WHERE s.id = po.original_student_id;

CREATE TEMP TABLE tmp_student_clone_map AS
WITH primary_owner AS (
    SELECT DISTINCT ON (original_student_id) original_student_id, user_owner_id
    FROM tmp_student_owner_map
    ORDER BY original_student_id, user_owner_id
)
SELECT gen_random_uuid() AS new_student_id, som.original_student_id, som.user_owner_id
FROM tmp_student_owner_map som
JOIN primary_owner po ON po.original_student_id = som.original_student_id
WHERE som.user_owner_id <> po.user_owner_id;

INSERT INTO students (
    id,
    student_id,
    name,
    phone,
    email,
    annotation,
    consent,
    created_at,
    updated_at,
    status,
    user_owner_id
)
SELECT
    scm.new_student_id,
    s.student_id,
    s.name,
    s.phone,
    s.email,
    s.annotation,
    s.consent,
    s.created_at,
    s.updated_at,
    s.status,
    scm.user_owner_id
FROM tmp_student_clone_map scm
JOIN students s ON s.id = scm.original_student_id;

UPDATE enrollments e
SET student_id = scm.new_student_id
FROM tmp_student_clone_map scm,
     disciplines d,
     programs p,
     campuses ca
WHERE e.student_id = scm.original_student_id
  AND d.id = e.discipline_id
  AND p.id = d.program_id
  AND ca.id = p.campus_id
  AND ca.user_owner_id = scm.user_owner_id;

ALTER TABLE students
DROP CONSTRAINT IF EXISTS students_student_id_key;

ALTER TABLE students
ADD CONSTRAINT students_user_owner_id_fkey
    FOREIGN KEY (user_owner_id) REFERENCES users(id) ON DELETE CASCADE;

CREATE UNIQUE INDEX students_user_owner_student_id_key
ON students (user_owner_id, student_id);
