-- Criar a função que atualiza o updated_at
CREATE OR REPLACE FUNCTION update_timestamp
()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'America/Sao_Paulo';
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Alterar users
ALTER TABLE users
    ALTER COLUMN created_at TYPE
TIMESTAMP
WITH TIME ZONE USING created_at AT TIME ZONE 'America/Sao_Paulo',
ALTER COLUMN updated_at TYPE TIMESTAMP
WITH TIME ZONE USING updated_at AT TIME ZONE 'America/Sao_Paulo';
CREATE TRIGGER trigger_update_timestamp_users
    BEFORE
UPDATE ON users
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp
();

-- Alterar whatsapp_instances
ALTER TABLE whatsapp_instances
    ALTER COLUMN created_at TYPE
TIMESTAMP
WITH TIME ZONE USING created_at AT TIME ZONE 'America/Sao_Paulo',
ALTER COLUMN updated_at TYPE TIMESTAMP
WITH TIME ZONE USING updated_at AT TIME ZONE 'America/Sao_Paulo';
CREATE TRIGGER trigger_update_timestamp_whatsapp_instances
    BEFORE
UPDATE ON whatsapp_instances
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp
();

-- Alterar smtp_instances
ALTER TABLE smtp_instances
    ALTER COLUMN created_at TYPE
TIMESTAMP
WITH TIME ZONE USING created_at AT TIME ZONE 'America/Sao_Paulo',
ALTER COLUMN updated_at TYPE TIMESTAMP
WITH TIME ZONE USING updated_at AT TIME ZONE 'America/Sao_Paulo';
CREATE TRIGGER trigger_update_timestamp_smtp_instances
    BEFORE
UPDATE ON smtp_instances
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp
();

-- Alterar campuses
ALTER TABLE campuses
    ALTER COLUMN created_at TYPE
TIMESTAMP
WITH TIME ZONE USING created_at AT TIME ZONE 'America/Sao_Paulo',
ALTER COLUMN updated_at TYPE TIMESTAMP
WITH TIME ZONE USING updated_at AT TIME ZONE 'America/Sao_Paulo';
CREATE TRIGGER trigger_update_timestamp_campuses
    BEFORE
UPDATE ON campuses
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp
();

-- Alterar programs
ALTER TABLE programs
    ALTER COLUMN created_at TYPE
TIMESTAMP
WITH TIME ZONE USING created_at AT TIME ZONE 'America/Sao_Paulo',
ALTER COLUMN updated_at TYPE TIMESTAMP
WITH TIME ZONE USING updated_at AT TIME ZONE 'America/Sao_Paulo';
CREATE TRIGGER trigger_update_timestamp_programs
    BEFORE
UPDATE ON programs
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp
();

-- Alterar courses
ALTER TABLE courses
    ALTER COLUMN created_at TYPE
TIMESTAMP
WITH TIME ZONE USING created_at AT TIME ZONE 'America/Sao_Paulo',
ALTER COLUMN updated_at TYPE TIMESTAMP
WITH TIME ZONE USING updated_at AT TIME ZONE 'America/Sao_Paulo';
CREATE TRIGGER trigger_update_timestamp_courses
    BEFORE
UPDATE ON courses
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp
();

-- Alterar students
ALTER TABLE students
    ALTER COLUMN created_at TYPE
TIMESTAMP
WITH TIME ZONE USING created_at AT TIME ZONE 'America/Sao_Paulo',
ALTER COLUMN updated_at TYPE TIMESTAMP
WITH TIME ZONE USING updated_at AT TIME ZONE 'America/Sao_Paulo';
CREATE TRIGGER trigger_update_timestamp_students
    BEFORE
UPDATE ON students
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp
();

-- Alterar enrollments
ALTER TABLE enrollments
    ALTER COLUMN created_at TYPE
TIMESTAMP
WITH TIME ZONE USING created_at AT TIME ZONE 'America/Sao_Paulo',
ALTER COLUMN updated_at TYPE TIMESTAMP
WITH TIME ZONE USING updated_at AT TIME ZONE 'America/Sao_Paulo';
CREATE TRIGGER trigger_update_timestamp_enrollments
    BEFORE
UPDATE ON enrollments
    FOR EACH ROW
EXECUTE FUNCTION update_timestamp
();