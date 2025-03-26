-- Reverter whatsapp_instances
ALTER TABLE whatsapp_instances
DROP CONSTRAINT whatsapp_instances_user_id_fkey;
ALTER TABLE whatsapp_instances
ADD CONSTRAINT whatsapp_instances_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE NO ACTION;

-- Reverter smtp_instances
ALTER TABLE smtp_instances
DROP CONSTRAINT smtp_instances_user_id_fkey;
ALTER TABLE smtp_instances
ADD CONSTRAINT smtp_instances_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE NO ACTION;

-- Reverter campuses
ALTER TABLE campuses
DROP CONSTRAINT campuses_user_owner_id_fkey;
ALTER TABLE campuses
ADD CONSTRAINT campuses_user_owner_id_fkey
    FOREIGN KEY (user_owner_id)
    REFERENCES users(id)
    ON DELETE NO ACTION;

-- Reverter programs
ALTER TABLE programs
DROP CONSTRAINT programs_campus_id_fkey;
ALTER TABLE programs
ADD CONSTRAINT programs_campus_id_fkey
    FOREIGN KEY (campus_id)
    REFERENCES campuses(id)
    ON DELETE NO ACTION;

-- Reverter courses
ALTER TABLE courses
DROP CONSTRAINT courses_program_id_fkey;
ALTER TABLE courses
ADD CONSTRAINT courses_program_id_fkey
    FOREIGN KEY (program_id)
    REFERENCES programs(id)
    ON DELETE NO ACTION;

-- Reverter enrollments
ALTER TABLE enrollments
DROP CONSTRAINT enrollments_course_id_fkey;
ALTER TABLE enrollments
DROP CONSTRAINT enrollments_student_id_fkey;
ALTER TABLE enrollments
ADD CONSTRAINT enrollments_course_id_fkey
    FOREIGN KEY (course_id)
    REFERENCES courses(id)
    ON DELETE NO ACTION;
ALTER TABLE enrollments
ADD CONSTRAINT enrollments_student_id_fkey
    FOREIGN KEY (student_id)
    REFERENCES students(id)
    ON DELETE NO ACTION;