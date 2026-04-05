package sqlite

import (
	"database/sql"
	"fmt"
)

const currentVersion = 1

// migration represents a schema migration step.
type migration struct {
	version    int
	statements []string
}

var migrations = []migration{
	{
		version: 1,
		statements: []string{
			`CREATE TABLE IF NOT EXISTS memories (
				id          INTEGER PRIMARY KEY AUTOINCREMENT,
				title       TEXT NOT NULL,
				type        TEXT NOT NULL,
				project     TEXT NOT NULL,
				scope       TEXT NOT NULL DEFAULT 'project',
				what        TEXT NOT NULL,
				why         TEXT NOT NULL,
				location    TEXT NOT NULL,
				learned     TEXT NOT NULL,
				tags        TEXT,
				topic_key   TEXT,
				created_at  TEXT NOT NULL DEFAULT (datetime('now')),
				updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
			)`,
			`CREATE INDEX IF NOT EXISTS idx_mem_project ON memories(project)`,
			`CREATE INDEX IF NOT EXISTS idx_mem_type ON memories(type)`,
			`CREATE INDEX IF NOT EXISTS idx_mem_scope ON memories(scope)`,
			`CREATE INDEX IF NOT EXISTS idx_mem_topic ON memories(topic_key, project, scope)`,
			`CREATE INDEX IF NOT EXISTS idx_mem_created ON memories(created_at DESC)`,
			`CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
				title, what, why, location, learned, tags,
				content='memories', content_rowid='id'
			)`,
			`CREATE TRIGGER IF NOT EXISTS mem_fts_insert AFTER INSERT ON memories BEGIN
				INSERT INTO memories_fts(rowid, title, what, why, location, learned, tags)
				VALUES (new.id, new.title, new.what, new.why, new.location, new.learned, new.tags);
			END`,
			`CREATE TRIGGER IF NOT EXISTS mem_fts_delete AFTER DELETE ON memories BEGIN
				INSERT INTO memories_fts(memories_fts, rowid, title, what, why, location, learned, tags)
				VALUES ('delete', old.id, old.title, old.what, old.why, old.location, old.learned, old.tags);
			END`,
			`CREATE TRIGGER IF NOT EXISTS mem_fts_update AFTER UPDATE ON memories BEGIN
				INSERT INTO memories_fts(memories_fts, rowid, title, what, why, location, learned, tags)
				VALUES ('delete', old.id, old.title, old.what, old.why, old.location, old.learned, old.tags);
				INSERT INTO memories_fts(rowid, title, what, why, location, learned, tags)
				VALUES (new.id, new.title, new.what, new.why, new.location, new.learned, new.tags);
			END`,
		},
	},
}

// getUserVersion returns the current schema version from the database.
func getUserVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("reading user_version: %w", err)
	}
	return version, nil
}

// setUserVersion sets the schema version in the database.
func setUserVersion(db *sql.DB, version int) error {
	_, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", version))
	if err != nil {
		return fmt.Errorf("setting user_version to %d: %w", version, err)
	}
	return nil
}

// runMigrations applies all pending migrations to the database.
func runMigrations(db *sql.DB) error {
	current, err := getUserVersion(db)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if m.version <= current {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("starting migration tx for v%d: %w", m.version, err)
		}

		for _, stmt := range m.statements {
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration v%d failed: %w", m.version, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration v%d: %w", m.version, err)
		}

		if err := setUserVersion(db, m.version); err != nil {
			return err
		}
	}

	return nil
}
