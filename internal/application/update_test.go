package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateMemoryUseCase_Execute(t *testing.T) {
	newTitle := "Updated title"
	invalidType := "invalid"
	validType := "decision"

	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockMemoryRepository
		args       func() UpdateMemoryRequest
		assert     func(t *testing.T, err error)
	}{
		{
			name: "updates with valid params",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Update(mock.Anything, int64(1), mock.AnythingOfType("domain.UpdateParams")).Return(nil)
				return m
			},
			args: func() UpdateMemoryRequest {
				return UpdateMemoryRequest{ID: 1, Title: &newTitle}
			},
			assert: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "converts string type to MemoryType",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Update(mock.Anything, int64(1), mock.MatchedBy(func(p domain.UpdateParams) bool {
					return p.Type != nil && *p.Type == domain.TypeDecision
				})).Return(nil)
				return m
			},
			args: func() UpdateMemoryRequest {
				return UpdateMemoryRequest{ID: 1, Type: &validType}
			},
			assert: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:       "rejects invalid type",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args: func() UpdateMemoryRequest {
				return UpdateMemoryRequest{ID: 1, Type: &invalidType}
			},
			assert: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidMemoryType)
			},
		},
		{
			name: "returns not found from repository",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Update(mock.Anything, int64(999), mock.Anything).Return(domain.ErrMemoryNotFound)
				return m
			},
			args: func() UpdateMemoryRequest {
				return UpdateMemoryRequest{ID: 999, Title: &newTitle}
			},
			assert: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrMemoryNotFound)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			req := tc.args()

			uc := NewUpdateMemoryUseCase(repo)
			err := uc.Execute(context.Background(), req)

			tc.assert(t, err)
			repo.AssertExpectations(t)
		})
	}
}
