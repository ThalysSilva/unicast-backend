-- Alterar whatsapp_instances
ALTER TABLE whatsapp_instances
DROP CONSTRAINT whatsapp_instances_user_id_fkey;
ALTER TABLE whatsapp_instances
ADD CONSTRAINT whatsapp_instances_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

-- Alterar smtp_instances
ALTER TABLE smtp_instances
DROP CONSTRAINT smtp_instances_user_id_fkey;
ALTER TABLE smtp_instances
ADD CONSTRAINT smtp_instances_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

-- Alterar campuses
ALTER TABLE campuses
DROP CONSTRAINT campuses_user_owner_id_fkey;
ALTER TABLE campuses
ADD CONSTRAINT campuses_user_owner_id_fkey
    FOREIGN KEY (user_owner_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

-- Alterar programs
ALTER TABLE programs
DROP CONSTRAINT programs_campus_id_fkey;
ALTER TABLE programs
ADD CONSTRAINT programs_campus_id_fkey
    FOREIGN KEY (campus_id)
    REFERENCES campuses(id)
    ON DELETE CASCADE;

-- Alterar courses
ALTER TABLE courses
DROP CONSTRAINT courses_program_id_fkey;
ALTER TABLE courses
ADD CONSTRAINT courses_program_id_fkey
    FOREIGN KEY (program_id)
    REFERENCES programs(id)
    ON DELETE CASCADE;

-- Alterar enrollments
ALTER TABLE enrollments
DROP CONSTRAINT enrollments_course_id_fkey;
ALTER TABLE enrollments
DROP CONSTRAINT enrollments_student_id_fkey;
ALTER TABLE enrollments
ADD CONSTRAINT enrollments_course_id_fkey
    FOREIGN KEY (course_id)
    REFERENCES courses(id)
    ON DELETE CASCADE;
ALTER TABLE enrollments
ADD CONSTRAINT enrollments_student_id_fkey
    FOREIGN KEY (student_id)
    REFERENCES students(id)
    ON DELETE CASCADE;