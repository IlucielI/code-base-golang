package migration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"code-base-golang/internal/config"
)

// PrintUsage outputs help instructions for database migration commands.
func PrintUsage() {
	fmt.Println(`Database Migration Tool

Usage:
  go run cmd/api/main.go migrate <command> [arguments]

Commands:
  create <name>     Create a new pair of .up.sql and .down.sql migration files
  up [n]            Apply all pending migrations (or next n steps)
  down [n|all]      Rollback last migration (or n steps, or "all")
  version, status   Show current schema version and dirty state
  force <version>   Force set database migration version (recovery tool)

Examples:
  go run cmd/api/main.go migrate create create_users_table
  go run cmd/api/main.go migrate up
  go run cmd/api/main.go migrate up 1
  go run cmd/api/main.go migrate down 1
  go run cmd/api/main.go migrate status
  go run cmd/api/main.go migrate force 20260928120000`)
}

// FindMigrationsDir locates the migrations directory relative to working directory or ancestors.
func FindMigrationsDir() string {
	// 1. Check explicit environment variable if set
	if envDir := os.Getenv("MIGRATIONS_DIR"); envDir != "" {
		if info, err := os.Stat(envDir); err == nil && info.IsDir() {
			abs, err := filepath.Abs(envDir)
			if err == nil {
				return abs
			}
			return envDir
		}
	}

	// 2. Search upwards from current working directory
	dir, err := os.Getwd()
	if err == nil {
		for {
			target := filepath.Join(dir, "migrations")
			if info, err := os.Stat(target); err == nil && info.IsDir() {
				return target
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return "migrations"
}

// Create generates a timestamped pair of up and down SQL migration files.
func Create(name string) (string, string, error) {
	sanitized := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	if sanitized == "" {
		return "", "", errors.New("migration name cannot be empty")
	}

	dir := FindMigrationsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create migrations directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	upPath := filepath.Join(dir, fmt.Sprintf("%s_%s.up.sql", timestamp, sanitized))
	downPath := filepath.Join(dir, fmt.Sprintf("%s_%s.down.sql", timestamp, sanitized))

	upContent := fmt.Sprintf("-- Migration: %s (UP)\n-- Created at: %s\n\n", name, time.Now().Format(time.RFC3339))
	downContent := fmt.Sprintf("-- Migration: %s (DOWN)\n-- Created at: %s\n\n", name, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write up migration: %w", err)
	}
	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write down migration: %w", err)
	}

	return upPath, downPath, nil
}

func initMigrate(cfg config.Config) (*migrate.Migrate, error) {
	dir := FindMigrationsDir()
	sourceURL := fmt.Sprintf("file://%s", dir)
	databaseURL := cfg.MigrationURI()

	return migrate.New(sourceURL, databaseURL)
}

// Up applies pending database migrations. Limit <= 0 applies all pending.
func Up(cfg config.Config, limit int) error {
	m, err := initMigrate(cfg)
	if err != nil {
		return fmt.Errorf("migration initialization failed: %w", err)
	}
	defer m.Close()

	if limit > 0 {
		fmt.Printf("Applying next %d migration(s)...\n", limit)
		err = m.Steps(limit)
	} else {
		fmt.Println("Applying all pending migrations...")
		err = m.Up()
	}

	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No new migrations to apply (database schema is already up to date).")
			return nil
		}
		return fmt.Errorf("migration UP failed: %w", err)
	}

	v, dirty, _ := m.Version()
	fmt.Printf("Successfully applied migrations. Current version: %d (dirty: %v)\n", v, dirty)
	return nil
}

// Down rolls back database migrations. Limit < 0 rolls back all.
func Down(cfg config.Config, limit int) error {
	m, err := initMigrate(cfg)
	if err != nil {
		return fmt.Errorf("migration initialization failed: %w", err)
	}
	defer m.Close()

	if limit < 0 {
		fmt.Println("Rolling back ALL migrations...")
		err = m.Down()
	} else {
		fmt.Printf("Rolling back %d migration step(s)...\n", limit)
		err = m.Steps(-limit)
	}

	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migrations to rollback.")
			return nil
		}
		return fmt.Errorf("migration DOWN failed: %w", err)
	}

	v, dirty, errVer := m.Version()
	if errors.Is(errVer, migrate.ErrNilVersion) {
		fmt.Println("All migrations rolled back. Schema is at clean initial state.")
	} else {
		fmt.Printf("Successfully rolled back. Current version: %d (dirty: %v)\n", v, dirty)
	}
	return nil
}

// Version prints current schema version and dirty flag.
func Version(cfg config.Config) error {
	m, err := initMigrate(cfg)
	if err != nil {
		return fmt.Errorf("migration initialization failed: %w", err)
	}
	defer m.Close()

	v, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("No migrations applied yet (version: 0)")
			return nil
		}
		return fmt.Errorf("failed to fetch migration version: %w", err)
	}

	fmt.Printf("Current Migration Version: %d\nDirty State: %v\n", v, dirty)
	return nil
}

// Force sets version state without running migration steps.
func Force(cfg config.Config, version int) error {
	m, err := initMigrate(cfg)
	if err != nil {
		return fmt.Errorf("migration initialization failed: %w", err)
	}
	defer m.Close()

	fmt.Printf("Forcing migration version to: %d...\n", version)
	if err := m.Force(version); err != nil {
		return fmt.Errorf("force migration version failed: %w", err)
	}

	fmt.Printf("Migration version successfully forced to: %d\n", version)
	return nil
}

// Run parses CLI migration sub-arguments and executes the corresponding action.
func Run(cfg config.Config, args []string) error {
	if len(args) == 0 {
		PrintUsage()
		return nil
	}

	command := strings.ToLower(args[0])

	switch command {
	case "create":
		if len(args) < 2 {
			return errors.New("migration name is required. Usage: go run cmd/api/main.go migrate create <name>")
		}
		name := strings.TrimSpace(args[1])
		up, down, err := Create(name)
		if err != nil {
			return err
		}
		fmt.Printf("Created migration files:\n  - %s\n  - %s\n", up, down)
		return nil

	case "up":
		limit := -1
		if len(args) >= 2 {
			n, err := strconv.Atoi(args[1])
			if err != nil || n <= 0 {
				return fmt.Errorf("invalid migration count: %s", args[1])
			}
			limit = n
		}
		return Up(cfg, limit)

	case "down":
		limit := 1
		if len(args) >= 2 {
			if strings.ToLower(args[1]) == "all" {
				limit = -1
			} else {
				n, err := strconv.Atoi(args[1])
				if err != nil || n <= 0 {
					return fmt.Errorf("invalid migration count: %s", args[1])
				}
				limit = n
			}
		}
		return Down(cfg, limit)

	case "version", "status":
		return Version(cfg)

	case "force":
		if len(args) < 2 {
			return errors.New("version argument required. Usage: go run cmd/api/main.go migrate force <version>")
		}
		v, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid version: %s", args[1])
		}
		return Force(cfg, v)

	default:
		PrintUsage()
		return fmt.Errorf("unknown migration command: %s", command)
	}
}
