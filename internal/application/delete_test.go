package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteMemoryUseCase_Execute(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		mockSetup func(*mocks.MockMemoryRepository)
		wantErr   error
	}{
		{
			name: "deletes existing memory",
			id:   1,
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Delete(mock.Anything, int64(1)).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "returns error for missing memory",
			id:   999,
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Delete(mock.Anything, int64(999)).Return(domain.ErrMemoryNotFound)
			},
			wantErr: domain.ErrMemoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMemoryRepository(t)
			tt.mockSetup(repo)

			uc := NewDeleteMemoryUseCase(repo)
			err := uc.Execute(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
