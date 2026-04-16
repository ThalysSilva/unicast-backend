DROP INDEX IF EXISTS smtp_instances_user_provider_email_oauth_key;

DELETE FROM smtp_instances WHERE auth_mode <> 'password';

ALTER TABLE smtp_instances
ALTER COLUMN password SET NOT NULL,
ALTER COLUMN iv SET NOT NULL;

ALTER TABLE smtp_instances
DROP COLUMN IF EXISTS token_expires_at,
DROP COLUMN IF EXISTS oauth_iv,
DROP COLUMN IF EXISTS oauth_payload,
DROP COLUMN IF EXISTS provider,
DROP COLUMN IF EXISTS auth_mode;
