package sqlite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) *Repository {
	t.Helper()
	db, err := OpenInMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewRepository(db)
}

func newTestMemory() *domain.Memory {
	return &domain.Memory{
		Title:    "Fixed auth bug",
		Type:     domain.TypeBugfix,
		Project:  "myapp",
		Scope:    domain.ScopeProject,
		What:     "Sanitized user input in query builder",
		Why:      "SQL injection vulnerability found in audit",
		Location: "src/db/query.go:142",
		Learned:  "Always use parameterized queries",
	}
}

func TestRepository_Save(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(t *testing.T) *Repository
		args   func() *domain.Memory
		assert func(t *testing.T, id int64, err error)
	}{
		{
			name:  "saves a basic memory",
			setup: func(t *testing.T) *Repository { return setupTestRepo(t) },
			args:  func() *domain.Memory { return newTestMemory() },
			assert: func(t *testing.T, id int64, err error) {
				require.NoError(t, err)
				assert.Greater(t, id, int64(0))
			},
		},
		{
			name:  "saves with tags",
			setup: func(t *testing.T) *Repository { return setupTestRepo(t) },
			args: func() *domain.Memory {
				m := newTestMemory()
				m.Tags = []string{"auth", "security"}
				return m
			},
			assert: func(t *testing.T, id int64, err error) {
				require.NoError(t, err)
				assert.Greater(t, id, int64(0))
			},
		},
		{
			name:  "saves with topic key",
			setup: func(t *testing.T) *Repository { return setupTestRepo(t) },
			args: func() *domain.Memory {
				m := newTestMemory()
				m.TopicKey = "architecture/auth-model"
				return m
			},
			assert: func(t *testing.T, id int64, err error) {
				require.NoError(t, err)
				assert.Greater(t, id, int64(0))
			},
		},
		{
			name:  "saves with personal scope",
			setup: func(t *testing.T) *Repository { return setupTestRepo(t) },
			args: func() *domain.Memory {
				m := newTestMemory()
				m.Scope = domain.ScopePersonal
				return m
			},
			assert: func(t *testing.T, id int64, err error) {
				require.NoError(t, err)
				assert.Greater(t, id, int64(0))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setup(t)
			memory := tc.args()

			id, err := repo.Save(context.Background(), memory)

			tc.assert(t, id, err)
		})
	}
}

func TestRepository_Save_TopicKeyUpsert(t *testing.T) {
	expectedTitle := "Updated auth decision"
	expectedWhat := "Switched to JWT"

	testCases := []struct {
		name   string
		setup  func(t *testing.T) (*Repository, int64)
		args   func() *domain.Memory
		assert func(t *testing.T, repo *Repository, originalID int64, upsertID int64, err error)
	}{
		{
			name: "upserts with same topic key returns same ID and updates content",
			setup: func(t *testing.T) (*Repository, int64) {
				repo := setupTestRepo(t)
				m := newTestMemory()
				m.TopicKey = "architecture/auth"
				id, err := repo.Save(context.Background(), m)
				require.NoError(t, err)
				return repo, id
			},
			args: func() *domain.Memory {
				m := newTestMemory()
				m.TopicKey = "architecture/auth"
				m.Title = expectedTitle
				m.What = expectedWhat
				return m
			},
			assert: func(t *testing.T, repo *Repository, originalID int64, upsertID int64, err error) {
				require.NoError(t, err)
				assert.Equal(t, originalID, upsertID)

				got, err := repo.GetByID(context.Background(), originalID)
				require.NoError(t, err)
				assert.Equal(t, expectedTitle, got.Title)
				assert.Equal(t, expectedWhat, got.What)
			},
		},
		{
			name: "different topic keys create different memories",
			setup: func(t *testing.T) (*Repository, int64) {
				repo := setupTestRepo(t)
				m := newTestMemory()
				m.TopicKey = "architecture/auth"
				id, err := repo.Save(context.Background(), m)
				require.NoError(t, err)
				return repo, id
			},
			args: func() *domain.Memory {
				m := newTestMemory()
				m.TopicKey = "architecture/db"
				return m
			},
			assert: func(t *testing.T, repo *Repository, originalID int64, newID int64, err error) {
				require.NoError(t, err)
				assert.NotEqual(t, originalID, newID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, originalID := tc.setup(t)
			memory := tc.args()

			upsertID, err := repo.Save(context.Background(), memory)

			tc.assert(t, repo, originalID, upsertID, err)
		})
	}
}

func TestRepository_GetByID(t *testing.T) {
	expectedTitle := "Fixed auth bug"

	testCases := []struct {
		name   string
		setup  func(t *testing.T) (*Repository, int64)
		assert func(t *testing.T, memory *domain.Memory, err error)
	}{
		{
			name: "retrieves existing memory",
			setup: func(t *testing.T) (*Repository, int64) {
				repo := setupTestRepo(t)
				id, err := repo.Save(context.Background(), newTestMemory())
				require.NoError(t, err)
				return repo, id
			},
			assert: func(t *testing.T, memory *domain.Memory, err error) {
				require.NoError(t, err)
				assert.Equal(t, expectedTitle, memory.Title)
			},
		},
		{
			name: "returns error for non-existent ID",
			setup: func(t *testing.T) (*Repository, int64) {
				repo := setupTestRepo(t)
				return repo, 9999
			},
			assert: func(t *testing.T, memory *domain.Memory, err error) {
				assert.ErrorIs(t, err, domain.ErrMemoryNotFound)
				assert.Nil(t, memory)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, id := tc.setup(t)

			memory, err := repo.GetByID(context.Background(), id)

			tc.assert(t, memory, err)
		})
	}
}

func TestRepository_Search(t *testing.T) {
	seedMemories := func(t *testing.T) *Repository {
		repo := setupTestRepo(t)
		memories := []*domain.Memory{
			{
				Title: "Fixed auth bug", Type: domain.TypeBugfix, Project: "myapp",
				Scope: domain.ScopeProject, What: "Sanitized input", Why: "SQL injection",
				Location: "src/db/query.go", Learned: "Use parameterized queries",
				Tags: []string{"auth", "security"},
			},
			{
				Title: "Database migration strategy", Type: domain.TypeDecision, Project: "myapp",
				Scope: domain.ScopeProject, What: "Chose goose for migrations", Why: "Simple and reliable",
				Location: "src/db/migrations/", Learned: "Goose supports both SQL and Go migrations",
			},
			{
				Title: "Docker setup", Type: domain.TypeConfig, Project: "other-project",
				Scope: domain.ScopeProject, What: "Added docker-compose", Why: "Local dev environment",
				Location: "docker-compose.yml", Learned: "Use named volumes for persistence",
			},
		}
		for _, m := range memories {
			_, err := repo.Save(context.Background(), m)
			require.NoError(t, err)
		}
		return repo
	}

	testCases := []struct {
		name   string
		setup  func(t *testing.T) *Repository
		args   func() domain.SearchQuery
		assert func(t *testing.T, results []domain.SearchResult, err error)
	}{
		{
			name:  "search by general term",
			setup: seedMemories,
			args:  func() domain.SearchQuery { return domain.SearchQuery{Text: "auth", Limit: 10} },
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				require.NoError(t, err)
				assert.Len(t, results, 1)
			},
		},
		{
			name:  "search cross-column",
			setup: seedMemories,
			args:  func() domain.SearchQuery { return domain.SearchQuery{Text: "injection", Limit: 10} },
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				require.NoError(t, err)
				assert.Len(t, results, 1)
			},
		},
		{
			name:  "search with project filter",
			setup: seedMemories,
			args: func() domain.SearchQuery {
				return domain.SearchQuery{Text: "migrations", Project: "myapp", Limit: 10}
			},
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				require.NoError(t, err)
				assert.Len(t, results, 1)
			},
		},
		{
			name:  "search with type filter",
			setup: seedMemories,
			args: func() domain.SearchQuery {
				return domain.SearchQuery{Text: "docker", Type: domain.TypeConfig, Limit: 10}
			},
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				require.NoError(t, err)
				assert.Len(t, results, 1)
			},
		},
		{
			name:  "search with no results",
			setup: seedMemories,
			args:  func() domain.SearchQuery { return domain.SearchQuery{Text: "nonexistent", Limit: 10} },
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				require.NoError(t, err)
				assert.Empty(t, results)
			},
		},
		{
			name:  "search by specific field tags",
			setup: seedMemories,
			args: func() domain.SearchQuery {
				return domain.SearchQuery{Text: "security", Field: "tags", Limit: 10}
			},
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				require.NoError(t, err)
				assert.Len(t, results, 1)
			},
		},
		{
			name:  "search by specific field location",
			setup: seedMemories,
			args: func() domain.SearchQuery {
				return domain.SearchQuery{Text: "docker-compose", Field: "location", Limit: 10}
			},
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				require.NoError(t, err)
				assert.Len(t, results, 1)
			},
		},
		{
			name:  "search respects limit",
			setup: seedMemories,
			args:  func() domain.SearchQuery { return domain.SearchQuery{Text: "src", Limit: 1} },
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				require.NoError(t, err)
				assert.Len(t, results, 1)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setup(t)
			query := tc.args()

			results, err := repo.Search(context.Background(), query)

			tc.assert(t, results, err)
		})
	}
}

func TestRepository_Update(t *testing.T) {
	expectedTitle := "Updated title"
	expectedWhat := "Updated what"
	expectedType := domain.TypeDecision

	testCases := []struct {
		name   string
		setup  func(t *testing.T) (*Repository, int64)
		args   func(id int64) (int64, domain.UpdateParams)
		assert func(t *testing.T, err error)
	}{
		{
			name: "update title",
			setup: func(t *testing.T) (*Repository, int64) {
				repo := setupTestRepo(t)
				id, err := repo.Save(context.Background(), newTestMemory())
				require.NoError(t, err)
				return repo, id
			},
			args: func(id int64) (int64, domain.UpdateParams) {
				title := expectedTitle
				return id, domain.UpdateParams{Title: &title}
			},
			assert: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "update multiple fields",
			setup: func(t *testing.T) (*Repository, int64) {
				repo := setupTestRepo(t)
				id, err := repo.Save(context.Background(), newTestMemory())
				require.NoError(t, err)
				return repo, id
			},
			args: func(id int64) (int64, domain.UpdateParams) {
				what := expectedWhat
				typ := expectedType
				return id, domain.UpdateParams{What: &what, Type: &typ}
			},
			assert: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "update non-existent memory",
			setup: func(t *testing.T) (*Repository, int64) {
				repo := setupTestRepo(t)
				return repo, 9999
			},
			args: func(id int64) (int64, domain.UpdateParams) {
				title := expectedTitle
				return id, domain.UpdateParams{Title: &title}
			},
			assert: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrMemoryNotFound)
			},
		},
		{
			name: "update with empty params is no-op",
			setup: func(t *testing.T) (*Repository, int64) {
				repo := setupTestRepo(t)
				id, err := repo.Save(context.Background(), newTestMemory())
				require.NoError(t, err)
				return repo, id
			},
			args: func(id int64) (int64, domain.UpdateParams) {
				return id, domain.UpdateParams{}
			},
			assert: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, id := tc.setup(t)
			targetID, params := tc.args(id)

			err := repo.Update(context.Background(), targetID, params)

			tc.assert(t, err)
		})
	}
}

func TestRepository_Delete(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(t *testing.T) (*Repository, int64)
		assert func(t *testing.T, repo *Repository, id int64, err error)
	}{
		{
			name: "delete existing memory",
			setup: func(t *testing.T) (*Repository, int64) {
				repo := setupTestRepo(t)
				id, err := repo.Save(context.Background(), newTestMemory())
				require.NoError(t, err)
				return repo, id
			},
			assert: func(t *testing.T, repo *Repository, id int64, err error) {
				require.NoError(t, err)
				_, getErr := repo.GetByID(context.Background(), id)
				assert.ErrorIs(t, getErr, domain.ErrMemoryNotFound)
			},
		},
		{
			name: "delete non-existent memory",
			setup: func(t *testing.T) (*Repository, int64) {
				repo := setupTestRepo(t)
				return repo, 9999
			},
			assert: func(t *testing.T, repo *Repository, id int64, err error) {
				assert.ErrorIs(t, err, domain.ErrMemoryNotFound)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, id := tc.setup(t)

			err := repo.Delete(context.Background(), id)

			tc.assert(t, repo, id, err)
		})
	}
}

func TestRepository_GetRecent(t *testing.T) {
	seedRepo := func(t *testing.T) *Repository {
		repo := setupTestRepo(t)
		for i, proj := range []string{"myapp", "myapp", "other"} {
			m := newTestMemory()
			m.Title = fmt.Sprintf("Memory %d", i+1)
			m.Project = proj
			if i == 0 {
				m.SessionID = "sess-1"
			}
			_, err := repo.Save(context.Background(), m)
			require.NoError(t, err)
		}
		return repo
	}

	testCases := []struct {
		name   string
		setup  func(t *testing.T) *Repository
		args   func() (string, string, int)
		assert func(t *testing.T, memories []domain.Memory, err error)
	}{
		{
			name:  "all recent memories",
			setup: seedRepo,
			args:  func() (string, string, int) { return "", "", 10 },
			assert: func(t *testing.T, memories []domain.Memory, err error) {
				require.NoError(t, err)
				assert.Len(t, memories, 3)
			},
		},
		{
			name:  "filtered by project",
			setup: seedRepo,
			args:  func() (string, string, int) { return "myapp", "", 10 },
			assert: func(t *testing.T, memories []domain.Memory, err error) {
				require.NoError(t, err)
				assert.Len(t, memories, 2)
			},
		},
		{
			name:  "filtered by session",
			setup: seedRepo,
			args:  func() (string, string, int) { return "", "sess-1", 10 },
			assert: func(t *testing.T, memories []domain.Memory, err error) {
				require.NoError(t, err)
				assert.Len(t, memories, 1)
				assert.Equal(t, "sess-1", memories[0].SessionID)
			},
		},
		{
			name:  "respects limit",
			setup: seedRepo,
			args:  func() (string, string, int) { return "", "", 1 },
			assert: func(t *testing.T, memories []domain.Memory, err error) {
				require.NoError(t, err)
				assert.Len(t, memories, 1)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setup(t)
			project, session, limit := tc.args()

			memories, err := repo.GetRecent(context.Background(), project, session, limit)

			tc.assert(t, memories, err)
		})
	}
}

func TestRepository_GetStats(t *testing.T) {
	seedRepo := func(t *testing.T) *Repository {
		repo := setupTestRepo(t)
		types := []domain.MemoryType{domain.TypeBugfix, domain.TypeBugfix, domain.TypeDecision}
		projects := []string{"app1", "app1", "app2"}
		for i := range types {
			m := newTestMemory()
			m.Type = types[i]
			m.Project = projects[i]
			_, err := repo.Save(context.Background(), m)
			require.NoError(t, err)
		}
		return repo
	}

	testCases := []struct {
		name   string
		setup  func(t *testing.T) *Repository
		args   func() string
		assert func(t *testing.T, stats *domain.Stats, err error)
	}{
		{
			name:  "empty database stats",
			setup: func(t *testing.T) *Repository { return setupTestRepo(t) },
			args:  func() string { return "" },
			assert: func(t *testing.T, stats *domain.Stats, err error) {
				require.NoError(t, err)
				assert.Equal(t, 0, stats.TotalMemories)
			},
		},
		{
			name:  "global stats",
			setup: seedRepo,
			args:  func() string { return "" },
			assert: func(t *testing.T, stats *domain.Stats, err error) {
				require.NoError(t, err)
				assert.Equal(t, 3, stats.TotalMemories)
				assert.Equal(t, 2, stats.ByType[domain.TypeBugfix])
				assert.Equal(t, 1, stats.ByType[domain.TypeDecision])
			},
		},
		{
			name:  "filtered by project",
			setup: seedRepo,
			args:  func() string { return "app1" },
			assert: func(t *testing.T, stats *domain.Stats, err error) {
				require.NoError(t, err)
				assert.Equal(t, 2, stats.TotalMemories)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setup(t)
			project := tc.args()

			stats, err := repo.GetStats(context.Background(), project)

			tc.assert(t, stats, err)
		})
	}
}

func TestRepository_GetAll(t *testing.T) {
	repo := setupTestRepo(t)

	// Seed data.
	for _, proj := range []string{"app1", "app1", "app2"} {
		m := newTestMemory()
		m.Project = proj
		_, err := repo.Save(context.Background(), m)
		require.NoError(t, err)
	}

	// All.
	all, err := repo.GetAll(context.Background(), "")
	require.NoError(t, err)
	assert.Len(t, all, 3)

	// Filtered.
	filtered, err := repo.GetAll(context.Background(), "app1")
	require.NoError(t, err)
	assert.Len(t, filtered, 2)

	// Empty.
	emptyRepo := setupTestRepo(t)
	empty, err := emptyRepo.GetAll(context.Background(), "")
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestRepository_SaveImport(t *testing.T) {
	repo := setupTestRepo(t)

	m := newTestMemory()
	m.CreatedAt = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	m.UpdatedAt = time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	m.SessionID = "imported-sess"
	m.Source = "import"

	id, err := repo.SaveImport(context.Background(), m)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Verify timestamps are preserved.
	got, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, "2025-01-01", got.CreatedAt.Format("2006-01-02"))
	assert.Equal(t, "2025-01-02", got.UpdatedAt.Format("2006-01-02"))
	assert.Equal(t, "imported-sess", got.SessionID)
	assert.Equal(t, "import", got.Source)
}
