package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchMemoryUseCase_Execute(t *testing.T) {
	tests := []struct {
		name      string
		req       SearchMemoryRequest
		mockSetup func(*mocks.MockMemoryRepository)
		wantCount int
		wantErr   error
	}{
		{
			name: "searches with text only",
			req:  SearchMemoryRequest{Text: "auth bug"},
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Search(mock.Anything, mock.MatchedBy(func(q domain.SearchQuery) bool {
					return q.Text == "auth bug" && q.Limit == 10
				})).Return([]domain.SearchResult{{}, {}}, nil)
			},
			wantCount: 2,
		},
		{
			name: "passes filters to repository",
			req: SearchMemoryRequest{
				Text:    "auth",
				Type:    "bugfix",
				Project: "myapp",
				Scope:   "project",
				Field:   "location",
				Limit:   5,
			},
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Search(mock.Anything, mock.MatchedBy(func(q domain.SearchQuery) bool {
					return q.Type == domain.TypeBugfix &&
						q.Project == "myapp" &&
						q.Scope == domain.ScopeProject &&
						q.Field == "location" &&
						q.Limit == 5
				})).Return([]domain.SearchResult{{}}, nil)
			},
			wantCount: 1,
		},
		{
			name:      "rejects empty text",
			req:       SearchMemoryRequest{Text: ""},
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrEmptySearchQuery,
		},
		{
			name:      "rejects invalid type",
			req:       SearchMemoryRequest{Text: "auth", Type: "invalid"},
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrInvalidMemoryType,
		},
		{
			name:      "rejects invalid scope",
			req:       SearchMemoryRequest{Text: "auth", Scope: "global"},
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrInvalidScope,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMemoryRepository(t)
			tt.mockSetup(repo)

			uc := NewSearchMemoryUseCase(repo)
			results, err := uc.Execute(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
		})
	}
}
