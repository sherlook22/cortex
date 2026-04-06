package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteMemoryUseCase_Execute(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockMemoryRepository
		args       func() int64
		assert     func(t *testing.T, err error)
	}{
		{
			name: "deletes existing memory",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Delete(mock.Anything, int64(1)).Return(nil)
				return m
			},
			args: func() int64 { return 1 },
			assert: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "returns error for missing memory",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Delete(mock.Anything, int64(999)).Return(domain.ErrMemoryNotFound)
				return m
			},
			args: func() int64 { return 999 },
			assert: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrMemoryNotFound)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			id := tc.args()

			uc := NewDeleteMemoryUseCase(repo)
			err := uc.Execute(context.Background(), id)

			tc.assert(t, err)
			repo.AssertExpectations(t)
		})
	}
}
