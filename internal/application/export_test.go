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

func TestExportUseCase_Execute(t *testing.T) {
	sampleMemories := []domain.Memory{
		{ID: 1, Title: "Memory one", Type: domain.TypeBugfix, Project: "myapp"},
		{ID: 2, Title: "Memory two", Type: domain.TypeDecision, Project: "myapp"},
	}

	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockMemoryRepository
		args       func() string
		assert     func(t *testing.T, data []byte, err error)
	}{
		{
			name: "exports all memories as JSON",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetAll(mock.Anything, "").Return(sampleMemories, nil)
				return m
			},
			args: func() string { return "" },
			assert: func(t *testing.T, data []byte, err error) {
				assert.NoError(t, err)
				var result []domain.Memory
				assert.NoError(t, json.Unmarshal(data, &result))
				assert.Len(t, result, 2)
				assert.Equal(t, "Memory one", result[0].Title)
			},
		},
		{
			name: "exports filtered by project",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetAll(mock.Anything, "myapp").Return(sampleMemories[:1], nil)
				return m
			},
			args: func() string { return "myapp" },
			assert: func(t *testing.T, data []byte, err error) {
				assert.NoError(t, err)
				var result []domain.Memory
				assert.NoError(t, json.Unmarshal(data, &result))
				assert.Len(t, result, 1)
			},
		},
		{
			name: "exports empty array for no memories",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetAll(mock.Anything, "").Return([]domain.Memory{}, nil)
				return m
			},
			args: func() string { return "" },
			assert: func(t *testing.T, data []byte, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "[]", string(data))
			},
		},
		{
			name: "returns error from repository",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetAll(mock.Anything, "").Return(nil, errors.New("db error"))
				return m
			},
			args: func() string { return "" },
			assert: func(t *testing.T, data []byte, err error) {
				assert.Error(t, err)
				assert.Nil(t, data)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			project := tc.args()

			uc := NewExportUseCase(repo)
			data, err := uc.Execute(context.Background(), project)

			tc.assert(t, data, err)
			repo.AssertExpectations(t)
		})
	}
}
