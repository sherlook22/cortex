package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetContextUseCase_Execute(t *testing.T) {
	sampleMemory := domain.Memory{
		ID: 1, Title: "Auth fix", Type: domain.TypeBugfix, Project: "myapp",
		What: "Fixed JWT", Why: "Token expired", Location: "src/auth.go",
		Learned: "Check expiry", CreatedAt: time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
	}

	memoryWithTags := domain.Memory{
		ID: 2, Title: "Config", Type: domain.TypeConfig, Project: "myapp",
		What: "Set DB pool", Why: "Perf", Location: "config.go",
		Learned: "Pool size matters", Tags: []string{"db", "perf"},
		CreatedAt: time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
	}

	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockMemoryRepository
		args       func() GetContextRequest
		assert     func(t *testing.T, result string, err error)
	}{
		{
			name: "returns formatted context",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetRecent(mock.Anything, "myapp", "", 10).Return([]domain.Memory{sampleMemory}, nil)
				return m
			},
			args: func() GetContextRequest { return GetContextRequest{Project: "myapp", Limit: 10} },
			assert: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, result, "Auth fix")
				assert.Contains(t, result, "Fixed JWT")
			},
		},
		{
			name: "returns empty for no memories",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetRecent(mock.Anything, "empty", "", 20).Return([]domain.Memory{}, nil)
				return m
			},
			args: func() GetContextRequest { return GetContextRequest{Project: "empty"} },
			assert: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Empty(t, result)
			},
		},
		{
			name: "defaults limit to 20",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetRecent(mock.Anything, "myapp", "", 20).Return([]domain.Memory{sampleMemory}, nil)
				return m
			},
			args: func() GetContextRequest { return GetContextRequest{Project: "myapp"} },
			assert: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			},
		},
		{
			name: "propagates repo error",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetRecent(mock.Anything, "myapp", "", 20).Return(nil, errors.New("db error"))
				return m
			},
			args: func() GetContextRequest { return GetContextRequest{Project: "myapp"} },
			assert: func(t *testing.T, result string, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "db error")
			},
		},
		{
			name: "formats context without project header",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetRecent(mock.Anything, "", "", 20).Return([]domain.Memory{sampleMemory}, nil)
				return m
			},
			args: func() GetContextRequest { return GetContextRequest{} },
			assert: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, result, "## Recent Memories\n")
			},
		},
		{
			name: "includes tags in context",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetRecent(mock.Anything, "myapp", "", 20).Return([]domain.Memory{memoryWithTags}, nil)
				return m
			},
			args: func() GetContextRequest { return GetContextRequest{Project: "myapp"} },
			assert: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.Contains(t, result, "db, perf")
			},
		},
		{
			name: "filters by session",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetRecent(mock.Anything, "myapp", "sess-1", 20).Return([]domain.Memory{sampleMemory}, nil)
				return m
			},
			args: func() GetContextRequest {
				return GetContextRequest{Project: "myapp", SessionID: "sess-1"}
			},
			assert: func(t *testing.T, result string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			req := tc.args()

			uc := NewGetContextUseCase(repo)
			result, err := uc.Execute(context.Background(), req)

			tc.assert(t, result, err)
			repo.AssertExpectations(t)
		})
	}
}
