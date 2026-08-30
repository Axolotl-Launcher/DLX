CREATE INDEX IF NOT EXISTS users_status_created_idx ON users(status, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS entitlements_status_user_idx ON entitlements(status, user_id);
CREATE INDEX IF NOT EXISTS afdian_orders_status_synced_idx ON afdian_orders(status, synced_at DESC);
CREATE INDEX IF NOT EXISTS afdian_orders_synced_idx ON afdian_orders(synced_at DESC);
CREATE INDEX IF NOT EXISTS usage_daily_date_user_idx ON usage_daily(date, user_id);
