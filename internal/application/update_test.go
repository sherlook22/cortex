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

	tests := []struct {
		name      string
		req       UpdateMemoryRequest
		mockSetup func(*mocks.MockMemoryRepository)
		wantErr   error
	}{
		{
			name: "updates with valid params",
			req:  UpdateMemoryRequest{ID: 1, Title: &newTitle},
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Update(mock.Anything, int64(1), mock.AnythingOfType("domain.UpdateParams")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "converts string type to MemoryType",
			req:  UpdateMemoryRequest{ID: 1, Type: &validType},
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Update(mock.Anything, int64(1), mock.MatchedBy(func(p domain.UpdateParams) bool {
					return p.Type != nil && *p.Type == domain.TypeDecision
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:      "rejects invalid type",
			req:       UpdateMemoryRequest{ID: 1, Type: &invalidType},
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrInvalidMemoryType,
		},
		{
			name: "returns not found from repository",
			req:  UpdateMemoryRequest{ID: 999, Title: &newTitle},
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Update(mock.Anything, int64(999), mock.Anything).Return(domain.ErrMemoryNotFound)
			},
			wantErr: domain.ErrMemoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMemoryRepository(t)
			tt.mockSetup(repo)

			uc := NewUpdateMemoryUseCase(repo)
			err := uc.Execute(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
