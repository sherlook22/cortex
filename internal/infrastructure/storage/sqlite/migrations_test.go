package sqlite

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(t *testing.T) *sql.DB
		assert func(t *testing.T, db *sql.DB)
	}{
		{
			name: "applies all migrations to fresh database",
			setup: func(t *testing.T) *sql.DB {
				db, err := OpenInMemory()
				require.NoError(t, err)
				return db
			},
			assert: func(t *testing.T, db *sql.DB) {
				version, err := getUserVersion(db)
				require.NoError(t, err)
				assert.Equal(t, currentVersion, version)
			},
		},
		{
			name: "migrations are idempotent",
			setup: func(t *testing.T) *sql.DB {
				db, err := OpenInMemory()
				require.NoError(t, err)
				err = runMigrations(db)
				require.NoError(t, err)
				return db
			},
			assert: func(t *testing.T, db *sql.DB) {
				version, err := getUserVersion(db)
				require.NoError(t, err)
				assert.Equal(t, currentVersion, version)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setup(t)
			defer db.Close()

			tc.assert(t, db)
		})
	}
}

func TestSchema_TablesExist(t *testing.T) {
	expectedTables := []string{"memories", "memories_fts", "sessions"}

	testCases := []struct {
		name   string
		setup  func(t *testing.T) *sql.DB
		args   func() string
		assert func(t *testing.T, db *sql.DB, table string)
	}{
		{
			name: "memories table exists",
			setup: func(t *testing.T) *sql.DB {
				db, err := OpenInMemory()
				require.NoError(t, err)
				return db
			},
			args: func() string { return expectedTables[0] },
			assert: func(t *testing.T, db *sql.DB, table string) {
				var name string
				err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
				require.NoError(t, err)
				assert.Equal(t, table, name)
			},
		},
		{
			name: "memories_fts virtual table exists",
			setup: func(t *testing.T) *sql.DB {
				db, err := OpenInMemory()
				require.NoError(t, err)
				return db
			},
			args: func() string { return expectedTables[1] },
			assert: func(t *testing.T, db *sql.DB, table string) {
				var name string
				err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
				require.NoError(t, err)
				assert.Equal(t, table, name)
			},
		},
		{
			name: "sessions table exists",
			setup: func(t *testing.T) *sql.DB {
				db, err := OpenInMemory()
				require.NoError(t, err)
				return db
			},
			args: func() string { return expectedTables[2] },
			assert: func(t *testing.T, db *sql.DB, table string) {
				var name string
				err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
				require.NoError(t, err)
				assert.Equal(t, table, name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setup(t)
			defer db.Close()

			table := tc.args()

			tc.assert(t, db, table)
		})
	}
}

func TestSchema_TriggersExist(t *testing.T) {
	expectedTriggers := []string{"mem_fts_insert", "mem_fts_delete", "mem_fts_update"}

	testCases := []struct {
		name   string
		setup  func(t *testing.T) *sql.DB
		args   func() string
		assert func(t *testing.T, db *sql.DB, trigger string)
	}{
		{
			name: "mem_fts_insert trigger exists",
			setup: func(t *testing.T) *sql.DB {
				db, err := OpenInMemory()
				require.NoError(t, err)
				return db
			},
			args: func() string { return expectedTriggers[0] },
			assert: func(t *testing.T, db *sql.DB, trigger string) {
				var name string
				err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='trigger' AND name=?", trigger).Scan(&name)
				require.NoError(t, err)
				assert.Equal(t, trigger, name)
			},
		},
		{
			name: "mem_fts_delete trigger exists",
			setup: func(t *testing.T) *sql.DB {
				db, err := OpenInMemory()
				require.NoError(t, err)
				return db
			},
			args: func() string { return expectedTriggers[1] },
			assert: func(t *testing.T, db *sql.DB, trigger string) {
				var name string
				err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='trigger' AND name=?", trigger).Scan(&name)
				require.NoError(t, err)
				assert.Equal(t, trigger, name)
			},
		},
		{
			name: "mem_fts_update trigger exists",
			setup: func(t *testing.T) *sql.DB {
				db, err := OpenInMemory()
				require.NoError(t, err)
				return db
			},
			args: func() string { return expectedTriggers[2] },
			assert: func(t *testing.T, db *sql.DB, trigger string) {
				var name string
				err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='trigger' AND name=?", trigger).Scan(&name)
				require.NoError(t, err)
				assert.Equal(t, trigger, name)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setup(t)
			defer db.Close()

			trigger := tc.args()

			tc.assert(t, db, trigger)
		})
	}
}

func TestSchema_IndexesExist(t *testing.T) {
	expectedIndexes := []string{
		"idx_mem_project", "idx_mem_type", "idx_mem_scope",
		"idx_mem_topic", "idx_mem_created",
		"idx_mem_session",
		"idx_sess_project", "idx_sess_status", "idx_sess_created",
	}

	testCases := []struct {
		name   string
		setup  func(t *testing.T) *sql.DB
		args   func() string
		assert func(t *testing.T, db *sql.DB, index string)
	}{}

	for _, idx := range expectedIndexes {
		idx := idx
		testCases = append(testCases, struct {
			name   string
			setup  func(t *testing.T) *sql.DB
			args   func() string
			assert func(t *testing.T, db *sql.DB, index string)
		}{
			name: idx + " exists",
			setup: func(t *testing.T) *sql.DB {
				db, err := OpenInMemory()
				require.NoError(t, err)
				return db
			},
			args: func() string { return idx },
			assert: func(t *testing.T, db *sql.DB, index string) {
				var name string
				err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", index).Scan(&name)
				require.NoError(t, err)
				assert.Equal(t, index, name)
			},
		})
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.setup(t)
			defer db.Close()

			index := tc.args()

			tc.assert(t, db, index)
		})
	}
}
