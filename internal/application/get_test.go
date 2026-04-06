package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMemoryUseCase_Execute(t *testing.T) {
	expectedTitle := "Found memory"

	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockMemoryRepository
		args       func() int64
		assert     func(t *testing.T, memory *domain.Memory, err error)
	}{
		{
			name: "retrieves existing memory",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetByID(mock.Anything, int64(1)).Return(&domain.Memory{
					ID: 1, Title: expectedTitle,
				}, nil)
				return m
			},
			args: func() int64 { return 1 },
			assert: func(t *testing.T, memory *domain.Memory, err error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedTitle, memory.Title)
			},
		},
		{
			name: "returns error for missing memory",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetByID(mock.Anything, int64(999)).Return(nil, domain.ErrMemoryNotFound)
				return m
			},
			args: func() int64 { return 999 },
			assert: func(t *testing.T, memory *domain.Memory, err error) {
				assert.ErrorIs(t, err, domain.ErrMemoryNotFound)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			id := tc.args()

			uc := NewGetMemoryUseCase(repo)
			memory, err := uc.Execute(context.Background(), id)

			tc.assert(t, memory, err)
			repo.AssertExpectations(t)
		})
	}
}
