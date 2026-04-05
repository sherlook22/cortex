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
	tests := []struct {
		name      string
		project   string
		mockSetup func(*mocks.MockMemoryRepository)
		wantTotal int
		wantErr   bool
	}{
		{
			name:    "returns global stats",
			project: "",
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().GetStats(mock.Anything, "").Return(&domain.Stats{
					TotalMemories: 5,
					ByType:        map[domain.MemoryType]int{domain.TypeBugfix: 3, domain.TypeDecision: 2},
					ByProject:     map[string]int{"app1": 3, "app2": 2},
				}, nil)
			},
			wantTotal: 5,
		},
		{
			name:    "returns project-filtered stats",
			project: "myapp",
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().GetStats(mock.Anything, "myapp").Return(&domain.Stats{
					TotalMemories: 3,
				}, nil)
			},
			wantTotal: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMemoryRepository(t)
			tt.mockSetup(repo)

			uc := NewGetStatsUseCase(repo)
			stats, err := uc.Execute(context.Background(), tt.project)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantTotal, stats.TotalMemories)
		})
	}
}
