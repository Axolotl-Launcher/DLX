ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS secret_ciphertext text;
