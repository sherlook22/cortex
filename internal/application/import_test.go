package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestImportUseCase_Execute(t *testing.T) {
	validMemories := []domain.Memory{
		{Title: "Memory one", Type: domain.TypeBugfix, Project: "myapp"},
		{Title: "Memory two", Type: domain.TypeDecision, Project: "myapp"},
	}
	validJSON, _ := json.Marshal(validMemories)

	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockMemoryRepository
		args       func() []byte
		assert     func(t *testing.T, count int, err error)
	}{
		{
			name: "imports all memories",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().SaveImport(mock.Anything, mock.AnythingOfType("*domain.Memory")).Return(int64(1), nil).Times(2)
				return m
			},
			args: func() []byte { return validJSON },
			assert: func(t *testing.T, count int, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 2, count)
			},
		},
		{
			name: "returns zero for empty array",
			setupMocks: func() *mocks.MockMemoryRepository {
				return mocks.NewMockMemoryRepository(t)
			},
			args: func() []byte { return []byte("[]") },
			assert: func(t *testing.T, count int, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 0, count)
			},
		},
		{
			name: "returns error for invalid JSON",
			setupMocks: func() *mocks.MockMemoryRepository {
				return mocks.NewMockMemoryRepository(t)
			},
			args: func() []byte { return []byte("not json") },
			assert: func(t *testing.T, count int, err error) {
				assert.Error(t, err)
				assert.Equal(t, 0, count)
			},
		},
		{
			name: "returns partial count on repository error",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().SaveImport(mock.Anything, mock.AnythingOfType("*domain.Memory")).
					Return(int64(1), nil).Once()
				m.EXPECT().SaveImport(mock.Anything, mock.AnythingOfType("*domain.Memory")).
					Return(int64(0), errors.New("db error")).Once()
				return m
			},
			args: func() []byte { return validJSON },
			assert: func(t *testing.T, count int, err error) {
				assert.Error(t, err)
				assert.Equal(t, 1, count)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			data := tc.args()

			uc := NewImportUseCase(repo)
			count, err := uc.Execute(context.Background(), data)

			tc.assert(t, count, err)
			repo.AssertExpectations(t)
		})
	}
}
