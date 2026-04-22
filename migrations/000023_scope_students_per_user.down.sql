DROP INDEX IF EXISTS students_user_owner_student_id_key;

ALTER TABLE students
DROP CONSTRAINT IF EXISTS students_user_owner_id_fkey;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM students
        GROUP BY student_id
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'cannot safely downgrade students ownership migration while duplicate student_id values exist';
    END IF;
END $$;

ALTER TABLE students
ADD CONSTRAINT students_student_id_key UNIQUE (student_id);

ALTER TABLE students
DROP COLUMN IF EXISTS user_owner_id;
