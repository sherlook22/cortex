package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const defaultDataDir = ".cortex"
const defaultDBName = "cortex.db"

// Config holds SQLite connection configuration.
type Config struct {
	// DataDir is the directory where the database file lives.
	// Defaults to ~/.cortex
	DataDir string
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		DataDir: filepath.Join(home, defaultDataDir),
	}
}

// Open creates or opens the SQLite database and runs pending migrations.
func Open(cfg Config) (*sql.DB, error) {
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data directory %s: %w", cfg.DataDir, err)
	}

	dbPath := filepath.Join(cfg.DataDir, defaultDBName)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database at %s: %w", dbPath, err)
	}

	// Configure SQLite for performance and safety.
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA foreign_keys = ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("setting pragma %q: %w", p, err)
		}
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}

// OpenInMemory creates an in-memory SQLite database for testing.
func OpenInMemory() (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("opening in-memory database: %w", err)
	}

	pragmas := []string{
		"PRAGMA foreign_keys = ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("setting pragma %q: %w", p, err)
		}
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}
