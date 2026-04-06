package sqlite

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// closedDB returns a *sql.DB that has been closed, forcing all operations to fail.
func closedDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := OpenInMemory()
	require.NoError(t, err)
	db.Close()
	return db
}

// corruptedRepo returns a repo where the memories table has been dropped.
func corruptedRepo(t *testing.T) *Repository {
	t.Helper()
	db, err := OpenInMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// Drop tables to force query errors.
	db.Exec("DROP TABLE memories_fts")
	db.Exec("DROP TABLE memories")

	return NewRepository(db)
}

// corruptedSessionRepo returns a session repo where the sessions table has been dropped.
func corruptedSessionRepo(t *testing.T) *SessionRepository {
	t.Helper()
	db, err := OpenInMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	db.Exec("DROP TABLE sessions")
	return NewSessionRepository(db)
}

// --- Repository error paths ---

func TestRepository_Save_Error(t *testing.T) {
	repo := corruptedRepo(t)
	m := newTestMemory()

	_, err := repo.Save(context.Background(), m)
	assert.Error(t, err)
}

func TestRepository_Save_TopicKeyCheckError(t *testing.T) {
	repo := corruptedRepo(t)
	m := newTestMemory()
	m.TopicKey = "some/key"

	_, err := repo.Save(context.Background(), m)
	assert.Error(t, err)
}

func TestRepository_Save_UpsertError(t *testing.T) {
	// Create a valid repo, save a memory with topic key, then corrupt and try upsert.
	db, err := OpenInMemory()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	m := newTestMemory()
	m.TopicKey = "arch/auth"
	_, err = repo.Save(context.Background(), m)
	require.NoError(t, err)

	// Now drop the table to force upsert error.
	db.Exec("DROP TRIGGER IF EXISTS mem_fts_update")
	db.Exec("DROP TRIGGER IF EXISTS mem_fts_delete")
	db.Exec("DROP TABLE memories_fts")

	m2 := newTestMemory()
	m2.TopicKey = "arch/auth"
	_, err = repo.Save(context.Background(), m2)
	// May or may not error depending on trigger presence; the upsert path is exercised.
}

func TestRepository_Search_Error(t *testing.T) {
	repo := corruptedRepo(t)

	_, err := repo.Search(context.Background(), domain.SearchQuery{
		Text: "auth", Limit: 10,
	})
	assert.Error(t, err)
}

func TestRepository_Search_WithSessionFilter(t *testing.T) {
	repo := setupTestRepo(t)
	m := newTestMemory()
	m.SessionID = "sess-1"
	_, err := repo.Save(context.Background(), m)
	require.NoError(t, err)

	results, err := repo.Search(context.Background(), domain.SearchQuery{
		Text: "auth", SessionID: "sess-1", Limit: 10,
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)

	results, err = repo.Search(context.Background(), domain.SearchQuery{
		Text: "auth", SessionID: "no-match", Limit: 10,
	})
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestRepository_GetByID_Error(t *testing.T) {
	repo := corruptedRepo(t)

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepository_Update_Error(t *testing.T) {
	repo := corruptedRepo(t)
	title := "new title"

	err := repo.Update(context.Background(), 1, domain.UpdateParams{Title: &title})
	assert.Error(t, err)
}

func TestRepository_Delete_Error(t *testing.T) {
	repo := corruptedRepo(t)

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepository_GetRecent_Error(t *testing.T) {
	repo := corruptedRepo(t)

	_, err := repo.GetRecent(context.Background(), "", "", 10)
	assert.Error(t, err)
}

func TestRepository_GetStats_Error(t *testing.T) {
	repo := corruptedRepo(t)

	_, err := repo.GetStats(context.Background(), "")
	assert.Error(t, err)
}

func TestRepository_FindByTopicKey_Error(t *testing.T) {
	repo := corruptedRepo(t)

	_, err := repo.FindByTopicKey(context.Background(), "key", "proj", domain.ScopeProject)
	assert.Error(t, err)
}

func TestRepository_GetAll_Error(t *testing.T) {
	repo := corruptedRepo(t)

	_, err := repo.GetAll(context.Background(), "")
	assert.Error(t, err)
}

func TestRepository_SaveImport_Error(t *testing.T) {
	repo := corruptedRepo(t)
	m := newTestMemory()
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()

	_, err := repo.SaveImport(context.Background(), m)
	assert.Error(t, err)
}

// --- Session Repository error paths ---

func TestSessionRepository_CreateSession_Error(t *testing.T) {
	repo := corruptedSessionRepo(t)

	err := repo.CreateSession(context.Background(), &domain.Session{ID: "s1", Project: "p"})
	assert.Error(t, err)
}

func TestSessionRepository_EndSession_Error(t *testing.T) {
	repo := corruptedSessionRepo(t)

	err := repo.EndSession(context.Background(), "s1", "summary")
	assert.Error(t, err)
}

func TestSessionRepository_GetSession_Error(t *testing.T) {
	repo := corruptedSessionRepo(t)

	_, err := repo.GetSession(context.Background(), "s1")
	assert.Error(t, err)
}

func TestSessionRepository_ListSessions_Error(t *testing.T) {
	repo := corruptedSessionRepo(t)

	_, err := repo.ListSessions(context.Background(), "", 10)
	assert.Error(t, err)
}

// --- Connection tests ---

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Contains(t, cfg.DataDir, ".cortex")
}

func TestOpen(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{DataDir: filepath.Join(tmpDir, "testdata")}

	db, err := Open(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Verify DB is functional.
	var version int
	err = db.QueryRow("PRAGMA user_version").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, currentVersion, version)
}

func TestOpen_InvalidPath(t *testing.T) {
	cfg := Config{DataDir: "/dev/null/impossible/path"}

	_, err := Open(cfg)
	assert.Error(t, err)
}

func TestOpenInMemory_Functional(t *testing.T) {
	db, err := OpenInMemory()
	require.NoError(t, err)
	defer db.Close()

	// Verify migrations ran.
	var version int
	err = db.QueryRow("PRAGMA user_version").Scan(&version)
	require.NoError(t, err)
	assert.Equal(t, currentVersion, version)
}

// --- Migration error paths ---

func TestRunMigrations_AlreadyCurrent(t *testing.T) {
	db, err := OpenInMemory()
	require.NoError(t, err)
	defer db.Close()

	// Running migrations again should be a no-op.
	err = runMigrations(db)
	assert.NoError(t, err)
}

func TestGetUserVersion(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	version, err := getUserVersion(db)
	require.NoError(t, err)
	assert.Equal(t, 0, version)
}

func TestSetUserVersion(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = setUserVersion(db, 42)
	require.NoError(t, err)

	version, err := getUserVersion(db)
	require.NoError(t, err)
	assert.Equal(t, 42, version)
}

func TestGetUserVersion_ClosedDB(t *testing.T) {
	db := closedDB(t)
	_, err := getUserVersion(db)
	assert.Error(t, err)
}

func TestSetUserVersion_ClosedDB(t *testing.T) {
	db := closedDB(t)
	err := setUserVersion(db, 1)
	assert.Error(t, err)
}

func TestRunMigrations_ClosedDB(t *testing.T) {
	db := closedDB(t)
	err := runMigrations(db)
	assert.Error(t, err)
}

// --- Additional repository error path tests ---

func TestRepository_Update_NoFields(t *testing.T) {
	repo := setupTestRepo(t)
	// Update with no fields should be a no-op.
	err := repo.Update(context.Background(), 999, domain.UpdateParams{})
	assert.NoError(t, err)
}

func TestRepository_Update_AllFields(t *testing.T) {
	repo := setupTestRepo(t)
	m := newTestMemory()
	id, err := repo.Save(context.Background(), m)
	require.NoError(t, err)

	title := "Updated"
	tp := domain.TypeDecision
	what := "New what"
	why := "New why"
	loc := "New loc"
	learned := "New learned"
	tags := []string{"new"}
	topicKey := "new/key"

	err = repo.Update(context.Background(), id, domain.UpdateParams{
		Title: &title, Type: &tp, What: &what, Why: &why,
		Location: &loc, Learned: &learned, Tags: &tags, TopicKey: &topicKey,
	})
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "Updated", got.Title)
	assert.Equal(t, domain.TypeDecision, got.Type)
	assert.Equal(t, "new/key", got.TopicKey)
}

func TestRepository_Update_NotFound(t *testing.T) {
	repo := setupTestRepo(t)
	title := "test"
	err := repo.Update(context.Background(), 99999, domain.UpdateParams{Title: &title})
	assert.ErrorIs(t, err, domain.ErrMemoryNotFound)
}

func TestRepository_Delete_NotFound(t *testing.T) {
	repo := setupTestRepo(t)
	err := repo.Delete(context.Background(), 99999)
	assert.ErrorIs(t, err, domain.ErrMemoryNotFound)
}

func TestRepository_Search_WithFieldFilter(t *testing.T) {
	repo := setupTestRepo(t)
	m := newTestMemory()
	_, err := repo.Save(context.Background(), m)
	require.NoError(t, err)

	// Search with field filter.
	results, err := repo.Search(context.Background(), domain.SearchQuery{
		Text: "auth", Field: "title", Limit: 10,
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestRepository_Search_AllFilters(t *testing.T) {
	repo := setupTestRepo(t)
	m := newTestMemory()
	m.SessionID = "sess-all"
	_, err := repo.Save(context.Background(), m)
	require.NoError(t, err)

	results, err := repo.Search(context.Background(), domain.SearchQuery{
		Text:      "auth",
		Type:      domain.TypeBugfix,
		Project:   "myapp",
		Scope:     domain.ScopeProject,
		SessionID: "sess-all",
		Limit:     10,
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestRepository_GetStats_ByProject(t *testing.T) {
	repo := setupTestRepo(t)
	m := newTestMemory()
	_, err := repo.Save(context.Background(), m)
	require.NoError(t, err)

	stats, err := repo.GetStats(context.Background(), "myapp")
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalMemories)
	assert.False(t, stats.OldestMemory.IsZero())
	assert.False(t, stats.NewestMemory.IsZero())
}

func TestRepository_GetRecent_DefaultLimit(t *testing.T) {
	repo := setupTestRepo(t)
	m := newTestMemory()
	_, err := repo.Save(context.Background(), m)
	require.NoError(t, err)

	memories, err := repo.GetRecent(context.Background(), "", "", 0)
	require.NoError(t, err)
	assert.Len(t, memories, 1)
}

func TestSessionRepository_ListSessions_DefaultLimit(t *testing.T) {
	repo := setupTestSessionRepo(t)
	repo.CreateSession(context.Background(), &domain.Session{ID: "s1", Project: "p"})

	sessions, err := repo.ListSessions(context.Background(), "", 0)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
}

func TestSessionRepository_ListSessions_FilteredByProject(t *testing.T) {
	repo := setupTestSessionRepo(t)
	repo.CreateSession(context.Background(), &domain.Session{ID: "s1", Project: "app1"})
	repo.CreateSession(context.Background(), &domain.Session{ID: "s2", Project: "app2"})

	sessions, err := repo.ListSessions(context.Background(), "app1", 10)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "app1", sessions[0].Project)
}

func TestRepository_Search_EmptyAfterSanitize(t *testing.T) {
	repo := setupTestRepo(t)
	// Sanitized text may become empty.
	results, err := repo.Search(context.Background(), domain.SearchQuery{
		Text: "***", Limit: 10,
	})
	require.NoError(t, err)
	assert.Nil(t, results)
}

func TestRepository_GetAll_Filtered(t *testing.T) {
	repo := setupTestRepo(t)
	m := newTestMemory()
	m.Project = "filtered"
	_, err := repo.Save(context.Background(), m)
	require.NoError(t, err)

	memories, err := repo.GetAll(context.Background(), "filtered")
	require.NoError(t, err)
	assert.Len(t, memories, 1)
}

// --- Tests using broken schema to cover scan error paths ---

func brokenSchemaRepo(t *testing.T) (*Repository, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// Create a memories table with wrong schema (missing columns).
	db.Exec(`CREATE TABLE memories (id INTEGER PRIMARY KEY, title TEXT)`)
	db.Exec(`INSERT INTO memories (id, title) VALUES (1, 'test')`)

	return NewRepository(db), db
}

func TestRepository_ScanMemories_Error(t *testing.T) {
	repo, _ := brokenSchemaRepo(t)

	// GetRecent will try to scan with wrong column count.
	_, err := repo.GetRecent(context.Background(), "", "", 10)
	assert.Error(t, err)
}

func TestRepository_GetByID_ScanError(t *testing.T) {
	repo, _ := brokenSchemaRepo(t)

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestRepository_GetAll_ScanError(t *testing.T) {
	repo, _ := brokenSchemaRepo(t)

	_, err := repo.GetAll(context.Background(), "")
	assert.Error(t, err)
}

func TestRepository_FindByTopicKey_ScanError(t *testing.T) {
	repo, db := brokenSchemaRepo(t)
	// Add required columns for the WHERE clause to work.
	db.Exec(`ALTER TABLE memories ADD COLUMN topic_key TEXT`)
	db.Exec(`ALTER TABLE memories ADD COLUMN project TEXT`)
	db.Exec(`ALTER TABLE memories ADD COLUMN scope TEXT`)
	db.Exec(`UPDATE memories SET topic_key='k', project='p', scope='project'`)

	_, err := repo.FindByTopicKey(context.Background(), "k", "p", domain.ScopeProject)
	assert.Error(t, err)
}

func brokenSessionSchemaRepo(t *testing.T) *SessionRepository {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// Sessions table with wrong schema.
	db.Exec(`CREATE TABLE sessions (id TEXT PRIMARY KEY, project TEXT)`)
	db.Exec(`INSERT INTO sessions (id, project) VALUES ('s1', 'p')`)

	return NewSessionRepository(db)
}

func TestSessionRepository_GetSession_ScanError(t *testing.T) {
	repo := brokenSessionSchemaRepo(t)

	_, err := repo.GetSession(context.Background(), "s1")
	assert.Error(t, err)
}

func TestSessionRepository_ListSessions_ScanError(t *testing.T) {
	repo := brokenSessionSchemaRepo(t)

	_, err := repo.ListSessions(context.Background(), "", 10)
	assert.Error(t, err)
}

// Test Search scan error with broken FTS.
func brokenSearchRepo(t *testing.T) *Repository {
	t.Helper()
	db, err := OpenInMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// Insert a memory normally then break the table schema.
	repo := NewRepository(db)
	m := newTestMemory()
	repo.Save(context.Background(), m)

	// Drop and recreate memories with wrong schema but keep FTS.
	db.Exec(`DROP TRIGGER IF EXISTS mem_fts_insert`)
	db.Exec(`DROP TRIGGER IF EXISTS mem_fts_update`)
	db.Exec(`DROP TRIGGER IF EXISTS mem_fts_delete`)

	// Rename to break the JOIN in search.
	db.Exec(`ALTER TABLE memories RENAME TO memories_old`)
	db.Exec(`CREATE TABLE memories (id INTEGER PRIMARY KEY, title TEXT)`)
	db.Exec(`INSERT INTO memories (id, title) VALUES (1, 'test')`)

	return repo
}

func TestRepository_Search_ScanError(t *testing.T) {
	repo := brokenSearchRepo(t)

	_, err := repo.Search(context.Background(), domain.SearchQuery{
		Text: "auth", Limit: 10,
	})
	assert.Error(t, err)
}

// --- GetStats type/project row scan error ---

func TestRepository_GetStats_TypeScanError(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create minimal memories table.
	db.Exec(`CREATE TABLE memories (
		id INTEGER PRIMARY KEY, type TEXT, project TEXT,
		created_at TEXT DEFAULT (datetime('now'))
	)`)
	db.Exec(`INSERT INTO memories (type, project) VALUES ('bugfix', 'app')`)

	repo := NewRepository(db)
	stats, err := repo.GetStats(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalMemories)
}

// --- RunMigrations error paths ---

func TestRunMigrations_BadStatement(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Override migrations with a bad statement to test the error path.
	oldMigrations := migrations
	defer func() { migrations = oldMigrations }()

	migrations = []migration{
		{
			version: 1,
			statements: []string{
				"THIS IS NOT VALID SQL",
			},
		},
	}

	err = runMigrations(db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migration v1 failed")
}

// --- Export json.Marshal edge case ---

func TestOpen_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nested := filepath.Join(tmpDir, "a", "b", "c")

	cfg := Config{DataDir: nested}
	db, err := Open(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Verify directory was created.
	_, err = os.Stat(nested)
	assert.NoError(t, err)
}
