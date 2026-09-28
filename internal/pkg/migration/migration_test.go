package migration_test

import (
	"os"
	"path/filepath"
	"testing"

	"code-base-golang/internal/config"
	"code-base-golang/internal/pkg/migration"
)

func TestFindMigrationsDir(t *testing.T) {
	dir := migration.FindMigrationsDir()
	if dir == "" {
		t.Fatal("expected non-empty migrations dir")
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatalf("expected migrations dir to exist and be directory, got %s (err: %v)", dir, err)
	}
}

func TestCreateMigrationFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "migrations-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	t.Setenv("MIGRATIONS_DIR", tmpDir)

	up, down, err := migration.Create("init_accounts")
	if err != nil {
		t.Fatalf("failed to create migration: %v", err)
	}
	defer os.Remove(up)
	defer os.Remove(down)

	if _, err := os.Stat(up); err != nil {
		t.Errorf("expected up file to exist: %v", err)
	}
	if _, err := os.Stat(down); err != nil {
		t.Errorf("expected down file to exist: %v", err)
	}

	if filepath.Ext(up) != ".sql" {
		t.Errorf("expected .sql extension, got %s", up)
	}
}

func TestRun_EmptyAndUnknown(t *testing.T) {
	cfg := config.Config{}

	// Empty args shows usage without error
	if err := migration.Run(cfg, []string{}); err != nil {
		t.Errorf("expected nil error on empty args, got %v", err)
	}

	// Unknown command returns error
	if err := migration.Run(cfg, []string{"unknown-cmd"}); err == nil {
		t.Error("expected error on unknown command, got nil")
	}
}
