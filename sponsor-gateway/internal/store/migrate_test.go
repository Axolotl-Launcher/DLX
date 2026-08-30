package store

import (
	"strings"
	"testing"
)

func TestEmbeddedInitialMigrationExists(t *testing.T) {
	source, err := migrationFiles.ReadFile("migrations/001_initial.sql")
	if err != nil {
		t.Fatal(err)
	}
	if len(source) == 0 {
		t.Fatal("initial migration was empty")
	}
}

func TestCDKMigrationContainsStorageAndSafetyConstraints(t *testing.T) {
	source, err := migrationFiles.ReadFile("migrations/004_cdk.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(source)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS cdk_batches",
		"CREATE TABLE IF NOT EXISTS cdks",
		"CREATE TABLE IF NOT EXISTS entitlement_ledger",
		"amount_fen bigint",
		"digest text NOT NULL UNIQUE",
		"status IN ('active', 'redeemed', 'revoked', 'expired')",
		"idempotency_key text NOT NULL UNIQUE",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("migration missing %q", required)
		}
	}
}

func TestAdminIndexesMigrationExists(t *testing.T) {
	source, err := migrationFiles.ReadFile("migrations/005_admin_indexes.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"users_status_created_idx", "entitlements_status_user_idx", "afdian_orders_status_synced_idx", "usage_daily_date_user_idx"} {
		if !strings.Contains(string(source), required) {
			t.Errorf("migration missing %q", required)
		}
	}
}
