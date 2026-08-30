-- CDK issuance and redemption are represented separately from provider orders.
-- Store only an HMAC digest of the code; plaintext CDKs must never reach SQL.
CREATE TABLE IF NOT EXISTS cdk_batches (
  id uuid PRIMARY KEY,
  name text NOT NULL,
  amount_fen bigint NOT NULL CHECK (amount_fen > 0),
  quantity integer NOT NULL CHECK (quantity > 0),
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cdks (
  id uuid PRIMARY KEY,
  batch_id uuid NOT NULL REFERENCES cdk_batches(id),
  digest text NOT NULL UNIQUE,
  amount_fen bigint NOT NULL CHECK (amount_fen > 0),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'redeemed', 'revoked', 'expired')),
  redeemed_by uuid REFERENCES users(id),
  redeemed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK ((status = 'redeemed') = (redeemed_by IS NOT NULL AND redeemed_at IS NOT NULL)),
  CHECK (status <> 'redeemed' OR redeemed_by IS NOT NULL),
  CHECK (status = 'redeemed' OR (redeemed_by IS NULL AND redeemed_at IS NULL))
);
CREATE INDEX IF NOT EXISTS cdks_batch_status_idx ON cdks(batch_id, status);

-- Immutable, idempotent balance movements. Positive entries grant fen; negative
-- entries consume/revoke fen. A CDK redemption references its source id.
CREATE TABLE IF NOT EXISTS entitlement_ledger (
  id bigserial PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id),
  amount_fen bigint NOT NULL CHECK (amount_fen <> 0),
  source_type text NOT NULL CHECK (source_type IN ('cdk', 'afdian', 'manual', 'adjustment')),
  source_id text NOT NULL,
  idempotency_key text NOT NULL UNIQUE,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS entitlement_ledger_user_created_idx ON entitlement_ledger(user_id, created_at, id);
