package sqlite

import (
	"testing"
)

func TestRunMigrations(t *testing.T) {
	tests := []struct {
		name            string
		expectedVersion int
		wantErr         bool
	}{
		{
			name:            "applies all migrations to fresh database",
			expectedVersion: currentVersion,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := OpenInMemory()
			if err != nil {
				t.Fatalf("OpenInMemory() error = %v", err)
			}
			defer db.Close()

			version, err := getUserVersion(db)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getUserVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
			if version != tt.expectedVersion {
				t.Errorf("getUserVersion() = %d, want %d", version, tt.expectedVersion)
			}
		})
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	db, err := OpenInMemory()
	if err != nil {
		t.Fatalf("first OpenInMemory() error = %v", err)
	}

	// Running migrations again should be a no-op.
	err = runMigrations(db)
	if err != nil {
		t.Fatalf("second runMigrations() error = %v", err)
	}

	version, err := getUserVersion(db)
	if err != nil {
		t.Fatalf("getUserVersion() error = %v", err)
	}
	if version != currentVersion {
		t.Errorf("getUserVersion() = %d, want %d", version, currentVersion)
	}
	db.Close()
}

func TestSchema_TablesExist(t *testing.T) {
	db, err := OpenInMemory()
	if err != nil {
		t.Fatalf("OpenInMemory() error = %v", err)
	}
	defer db.Close()

	tables := []struct {
		name     string
		query    string
		wantRows bool
	}{
		{
			name:     "memories table exists",
			query:    "SELECT name FROM sqlite_master WHERE type='table' AND name='memories'",
			wantRows: true,
		},
		{
			name:     "memories_fts virtual table exists",
			query:    "SELECT name FROM sqlite_master WHERE type='table' AND name='memories_fts'",
			wantRows: true,
		},
	}

	for _, tt := range tables {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.query)
			if err != nil {
				t.Fatalf("query error: %v", err)
			}
			defer rows.Close()

			hasRows := rows.Next()
			if hasRows != tt.wantRows {
				t.Errorf("table check: got rows=%v, want %v", hasRows, tt.wantRows)
			}
		})
	}
}

func TestSchema_TriggersExist(t *testing.T) {
	db, err := OpenInMemory()
	if err != nil {
		t.Fatalf("OpenInMemory() error = %v", err)
	}
	defer db.Close()

	triggers := []string{"mem_fts_insert", "mem_fts_delete", "mem_fts_update"}

	for _, trigger := range triggers {
		t.Run(trigger, func(t *testing.T) {
			var name string
			err := db.QueryRow(
				"SELECT name FROM sqlite_master WHERE type='trigger' AND name=?",
				trigger,
			).Scan(&name)
			if err != nil {
				t.Errorf("trigger %q not found: %v", trigger, err)
			}
		})
	}
}

func TestSchema_IndexesExist(t *testing.T) {
	db, err := OpenInMemory()
	if err != nil {
		t.Fatalf("OpenInMemory() error = %v", err)
	}
	defer db.Close()

	indexes := []string{
		"idx_mem_project",
		"idx_mem_type",
		"idx_mem_scope",
		"idx_mem_topic",
		"idx_mem_created",
	}

	for _, idx := range indexes {
		t.Run(idx, func(t *testing.T) {
			var name string
			err := db.QueryRow(
				"SELECT name FROM sqlite_master WHERE type='index' AND name=?",
				idx,
			).Scan(&name)
			if err != nil {
				t.Errorf("index %q not found: %v", idx, err)
			}
		})
	}
}
