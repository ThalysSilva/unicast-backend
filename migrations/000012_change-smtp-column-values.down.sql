ALTER TABLE smtp_instances
    ALTER COLUMN password TYPE TEXT USING encode(password, 'escape'),
    ALTER COLUMN iv TYPE TEXT USING encode(iv, 'escape');