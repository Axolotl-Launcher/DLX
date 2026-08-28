package store

import "testing"

func TestEmbeddedInitialMigrationExists(t *testing.T) {
	source, err := migrationFiles.ReadFile("migrations/001_initial.sql")
	if err != nil {
		t.Fatal(err)
	}
	if len(source) == 0 {
		t.Fatal("initial migration was empty")
	}
}
