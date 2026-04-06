package sqlite

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestSessionRepo(t *testing.T) *SessionRepository {
	t.Helper()
	db, err := OpenInMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewSessionRepository(db)
}

func TestSessionRepository_CreateSession(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(t *testing.T) *SessionRepository
		args   *domain.Session
		assert func(t *testing.T, repo *SessionRepository, err error)
	}{
		{
			name:  "creates session",
			setup: setupTestSessionRepo,
			args:  &domain.Session{ID: "s1", Project: "myapp", Directory: "/home/dev"},
			assert: func(t *testing.T, repo *SessionRepository, err error) {
				require.NoError(t, err)
				s, err := repo.GetSession(context.Background(), "s1")
				require.NoError(t, err)
				assert.Equal(t, "s1", s.ID)
				assert.Equal(t, "myapp", s.Project)
				assert.Equal(t, domain.SessionActive, s.Status)
			},
		},
		{
			name: "idempotent create",
			setup: func(t *testing.T) *SessionRepository {
				repo := setupTestSessionRepo(t)
				repo.CreateSession(context.Background(), &domain.Session{ID: "s1", Project: "myapp"})
				return repo
			},
			args: &domain.Session{ID: "s1", Project: "other"},
			assert: func(t *testing.T, repo *SessionRepository, err error) {
				require.NoError(t, err)
				// Original project is preserved (INSERT OR IGNORE).
				s, _ := repo.GetSession(context.Background(), "s1")
				assert.Equal(t, "myapp", s.Project)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setup(t)
			err := repo.CreateSession(context.Background(), tc.args)
			tc.assert(t, repo, err)
		})
	}
}

func TestSessionRepository_EndSession(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(t *testing.T) *SessionRepository
		id     string
		assert func(t *testing.T, repo *SessionRepository, err error)
	}{
		{
			name: "ends session with summary",
			setup: func(t *testing.T) *SessionRepository {
				repo := setupTestSessionRepo(t)
				repo.CreateSession(context.Background(), &domain.Session{ID: "s1", Project: "myapp"})
				return repo
			},
			id: "s1",
			assert: func(t *testing.T, repo *SessionRepository, err error) {
				require.NoError(t, err)
				s, _ := repo.GetSession(context.Background(), "s1")
				assert.Equal(t, domain.SessionCompleted, s.Status)
				assert.Equal(t, "done", s.Summary)
			},
		},
		{
			name:  "returns error for non-existent session",
			setup: setupTestSessionRepo,
			id:    "no-exist",
			assert: func(t *testing.T, repo *SessionRepository, err error) {
				assert.ErrorIs(t, err, domain.ErrSessionNotFound)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setup(t)
			err := repo.EndSession(context.Background(), tc.id, "done")
			tc.assert(t, repo, err)
		})
	}
}

func TestSessionRepository_GetSession(t *testing.T) {
	repo := setupTestSessionRepo(t)
	repo.CreateSession(context.Background(), &domain.Session{ID: "s1", Project: "myapp", Directory: "/dev"})

	s, err := repo.GetSession(context.Background(), "s1")
	require.NoError(t, err)
	assert.Equal(t, "s1", s.ID)
	assert.Equal(t, "myapp", s.Project)
	assert.Equal(t, "/dev", s.Directory)

	_, err = repo.GetSession(context.Background(), "no-exist")
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)
}

func TestSessionRepository_ListSessions(t *testing.T) {
	repo := setupTestSessionRepo(t)
	repo.CreateSession(context.Background(), &domain.Session{ID: "s1", Project: "app1"})
	repo.CreateSession(context.Background(), &domain.Session{ID: "s2", Project: "app1"})
	repo.CreateSession(context.Background(), &domain.Session{ID: "s3", Project: "app2"})

	// All sessions.
	all, err := repo.ListSessions(context.Background(), "", 10)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	// Filtered by project.
	filtered, err := repo.ListSessions(context.Background(), "app1", 10)
	require.NoError(t, err)
	assert.Len(t, filtered, 2)

	// Respects limit.
	limited, err := repo.ListSessions(context.Background(), "", 1)
	require.NoError(t, err)
	assert.Len(t, limited, 1)
}
