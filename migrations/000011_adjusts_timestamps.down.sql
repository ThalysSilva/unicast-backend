-- Remover triggers
DROP TRIGGER IF EXISTS trigger_update_timestamp_users
ON users;
DROP TRIGGER IF EXISTS trigger_update_timestamp_whatsapp_instances
ON whatsapp_instances;
DROP TRIGGER IF EXISTS trigger_update_timestamp_smtp_instances
ON smtp_instances;
DROP TRIGGER IF EXISTS trigger_update_timestamp_campuses
ON campuses;
DROP TRIGGER IF EXISTS trigger_update_timestamp_programs
ON programs;
DROP TRIGGER IF EXISTS trigger_update_timestamp_courses
ON courses;
DROP TRIGGER IF EXISTS trigger_update_timestamp_students
ON students;
DROP TRIGGER IF EXISTS trigger_update_timestamp_enrollments
ON enrollments;

-- Dropar a função
DROP FUNCTION IF EXISTS update_timestamp;

-- Reverter users
ALTER TABLE users
    ALTER COLUMN created_at TYPE
TIMESTAMP USING created_at AT TIME ZONE 'UTC',
ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

-- Reverter whatsapp_instances
ALTER TABLE whatsapp_instances
    ALTER COLUMN created_at TYPE
TIMESTAMP USING created_at AT TIME ZONE 'UTC',
ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

-- Reverter smtp_instances
ALTER TABLE smtp_instances
    ALTER COLUMN created_at TYPE
TIMESTAMP USING created_at AT TIME ZONE 'UTC',
ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

-- Reverter campuses
ALTER TABLE campuses
    ALTER COLUMN created_at TYPE
TIMESTAMP USING created_at AT TIME ZONE 'UTC',
ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

-- Reverter programs
ALTER TABLE programs
    ALTER COLUMN created_at TYPE
TIMESTAMP USING created_at AT TIME ZONE 'UTC',
ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

-- Reverter courses
ALTER TABLE courses
    ALTER COLUMN created_at TYPE
TIMESTAMP USING created_at AT TIME ZONE 'UTC',
ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

-- Reverter students
ALTER TABLE students
    ALTER COLUMN created_at TYPE
TIMESTAMP USING created_at AT TIME ZONE 'UTC',
ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

-- Reverter enrollments
ALTER TABLE enrollments
    ALTER COLUMN created_at TYPE
TIMESTAMP USING created_at AT TIME ZONE 'UTC',
ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';