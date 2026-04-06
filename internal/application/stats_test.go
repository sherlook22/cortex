package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetStatsUseCase_Execute(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockMemoryRepository
		args       func() string
		assert     func(t *testing.T, stats *domain.Stats, err error)
	}{
		{
			name: "returns global stats",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetStats(mock.Anything, "").Return(&domain.Stats{
					TotalMemories: 5,
					ByType:        map[domain.MemoryType]int{domain.TypeBugfix: 3, domain.TypeDecision: 2},
					ByProject:     map[string]int{"app1": 3, "app2": 2},
				}, nil)
				return m
			},
			args: func() string { return "" },
			assert: func(t *testing.T, stats *domain.Stats, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 5, stats.TotalMemories)
			},
		},
		{
			name: "returns project-filtered stats",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().GetStats(mock.Anything, "myapp").Return(&domain.Stats{
					TotalMemories: 3,
				}, nil)
				return m
			},
			args: func() string { return "myapp" },
			assert: func(t *testing.T, stats *domain.Stats, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 3, stats.TotalMemories)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			project := tc.args()

			uc := NewGetStatsUseCase(repo)
			stats, err := uc.Execute(context.Background(), project)

			tc.assert(t, stats, err)
			repo.AssertExpectations(t)
		})
	}
}
