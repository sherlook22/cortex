package application

import (
	"context"
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
				m.EXPECT().GetRecent(mock.Anything, "myapp", 10).Return([]domain.Memory{sampleMemory}, nil)
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
				m.EXPECT().GetRecent(mock.Anything, "empty", 20).Return([]domain.Memory{}, nil)
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
				m.EXPECT().GetRecent(mock.Anything, "myapp", 20).Return([]domain.Memory{sampleMemory}, nil)
				return m
			},
			args: func() GetContextRequest { return GetContextRequest{Project: "myapp"} },
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
