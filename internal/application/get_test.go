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
	tests := []struct {
		name      string
		id        int64
		mockSetup func(*mocks.MockMemoryRepository)
		wantTitle string
		wantErr   error
	}{
		{
			name: "retrieves existing memory",
			id:   1,
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().GetByID(mock.Anything, int64(1)).Return(&domain.Memory{
					ID:    1,
					Title: "Found memory",
				}, nil)
			},
			wantTitle: "Found memory",
		},
		{
			name: "returns error for missing memory",
			id:   999,
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().GetByID(mock.Anything, int64(999)).Return(nil, domain.ErrMemoryNotFound)
			},
			wantErr: domain.ErrMemoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMemoryRepository(t)
			tt.mockSetup(repo)

			uc := NewGetMemoryUseCase(repo)
			got, err := uc.Execute(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantTitle, got.Title)
		})
	}
}
