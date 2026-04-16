ALTER TABLE smtp_instances
ADD COLUMN auth_mode VARCHAR NOT NULL DEFAULT 'password',
ADD COLUMN provider VARCHAR NOT NULL DEFAULT 'custom_smtp',
ADD COLUMN oauth_payload BYTEA NULL,
ADD COLUMN oauth_iv BYTEA NULL,
ADD COLUMN token_expires_at TIMESTAMP NULL;

ALTER TABLE smtp_instances
ALTER COLUMN password DROP NOT NULL,
ALTER COLUMN iv DROP NOT NULL;

CREATE UNIQUE INDEX smtp_instances_user_provider_email_oauth_key
ON smtp_instances(user_id, provider, email)
WHERE auth_mode = 'oauth';
