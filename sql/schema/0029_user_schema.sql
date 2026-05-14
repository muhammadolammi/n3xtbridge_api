-- +goose Up
-- 1. Make password and profile fields nullable for OAuth users
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ALTER COLUMN address DROP NOT NULL;
ALTER TABLE users ALTER COLUMN country DROP NOT NULL;
ALTER TABLE users ALTER COLUMN state DROP NOT NULL;

-- 2. Add verification and OAuth tracking
ALTER TABLE users ADD COLUMN is_email_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN google_id VARCHAR(255) UNIQUE; -- Best way to identify Google users
ALTER TABLE users ADD COLUMN avatar_url TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN avatar_url;
ALTER TABLE users DROP COLUMN google_id;
ALTER TABLE users DROP COLUMN is_email_verified;
