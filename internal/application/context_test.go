package application

import (
	"context"
	"testing"
	"time"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetContextUseCase_Execute(t *testing.T) {
	sampleMemory := domain.Memory{
		ID:        1,
		Title:     "Auth fix",
		Type:      domain.TypeBugfix,
		Project:   "myapp",
		What:      "Fixed JWT",
		Why:       "Token expired",
		Location:  "src/auth.go",
		Learned:   "Check expiry",
		CreatedAt: time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name      string
		req       GetContextRequest
		mockSetup func(*mocks.MockMemoryRepository)
		wantEmpty bool
		wantErr   bool
	}{
		{
			name: "returns formatted context",
			req:  GetContextRequest{Project: "myapp", Limit: 10},
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().GetRecent(mock.Anything, "myapp", 10).Return([]domain.Memory{sampleMemory}, nil)
			},
			wantEmpty: false,
		},
		{
			name: "returns empty for no memories",
			req:  GetContextRequest{Project: "empty"},
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().GetRecent(mock.Anything, "empty", 20).Return([]domain.Memory{}, nil)
			},
			wantEmpty: true,
		},
		{
			name: "defaults limit to 20",
			req:  GetContextRequest{Project: "myapp"},
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().GetRecent(mock.Anything, "myapp", 20).Return([]domain.Memory{sampleMemory}, nil)
			},
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMemoryRepository(t)
			tt.mockSetup(repo)

			uc := NewGetContextUseCase(repo)
			result, err := uc.Execute(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.wantEmpty {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, "Auth fix")
				assert.Contains(t, result, "Fixed JWT")
			}
		})
	}
}
