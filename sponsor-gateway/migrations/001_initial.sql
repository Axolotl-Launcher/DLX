-- Migration inputs are server-validated and minimized. Never store translation text,
-- bearer credentials, Afdian tokens, or unredacted address/contact data in JSON columns.
CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY,
  email text NOT NULL UNIQUE,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'blocked')),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS afdian_identities (
  user_id uuid PRIMARY KEY REFERENCES users(id),
  afdian_user_id text NOT NULL UNIQUE,
  verified_at timestamptz NOT NULL
);
CREATE TABLE IF NOT EXISTS afdian_orders (
  out_trade_no text PRIMARY KEY,
  afdian_user_id text NOT NULL,
  actual_paid_fen bigint NOT NULL CHECK (actual_paid_fen >= 0),
  status text NOT NULL CHECK (status IN ('paid', 'success', 'pending', 'refunded', 'revoked', 'cancelled', 'unknown')),
  -- Redacted provider fields only; application policy deletes this after retention period.
  raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  synced_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS afdian_orders_identity_idx ON afdian_orders(afdian_user_id);
CREATE TABLE IF NOT EXISTS entitlements (
  user_id uuid PRIMARY KEY REFERENCES users(id),
  lifetime_paid_fen bigint NOT NULL DEFAULT 0 CHECK (lifetime_paid_fen >= 0),
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'granted', 'suspended', 'manual_review')),
  granted_at timestamptz,
  recalculated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS api_keys (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id),
  -- Prefix is an opaque key id, not a recognizable secret fragment.
  prefix text NOT NULL,
  secret_hash text NOT NULL,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked', 'suspended')),
  created_at timestamptz NOT NULL DEFAULT now(),
  last_used_at timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS one_active_key_per_user ON api_keys(user_id) WHERE status = 'active';
CREATE TABLE IF NOT EXISTS usage_daily (
  user_id uuid NOT NULL REFERENCES users(id), date date NOT NULL,
  request_count bigint NOT NULL DEFAULT 0 CHECK (request_count >= 0),
  input_chars bigint NOT NULL DEFAULT 0 CHECK (input_chars >= 0),
  error_count bigint NOT NULL DEFAULT 0 CHECK (error_count >= 0),
  PRIMARY KEY(user_id,date)
);
CREATE TABLE IF NOT EXISTS webhook_events (
  provider_event_key text PRIMARY KEY,
  -- This must be a redacted/minimized event envelope, not raw personal data.
  payload jsonb NOT NULL,
  received_at timestamptz NOT NULL DEFAULT now(), processed_at timestamptz,
  result text
);
CREATE TABLE IF NOT EXISTS audit_logs (
  id bigserial PRIMARY KEY, actor text NOT NULL, action text NOT NULL, target text NOT NULL,
  -- metadata excludes emails, keys, text, tokens, and raw payment payloads.
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb, created_at timestamptz NOT NULL DEFAULT now()
);
