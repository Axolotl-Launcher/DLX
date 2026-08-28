-- Recovery login codes replace email verification in this deployment.
-- The plaintext code is never stored; it is shown exactly once after verified claim.
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS login_code_hash text;
CREATE UNIQUE INDEX IF NOT EXISTS users_login_code_hash_unique ON users(login_code_hash) WHERE login_code_hash IS NOT NULL;
