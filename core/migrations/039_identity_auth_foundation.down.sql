-- Migration 039 down: remove V8.2 identity/auth foundation objects.

DROP TABLE IF EXISTS identity_audit_events;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS auth_providers;
DROP TABLE IF EXISTS org_memberships;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS groups;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_account_id_fkey'
    ) THEN
        ALTER TABLE users DROP CONSTRAINT users_account_id_fkey;
    END IF;
END $$;

DROP INDEX IF EXISTS idx_users_account_status;
DROP INDEX IF EXISTS idx_users_account_email_unique;
DROP INDEX IF EXISTS idx_users_account_username_unique;

ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_at,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS external_subject,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS display_name,
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS account_id;

DROP TABLE IF EXISTS accounts;
