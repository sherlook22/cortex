package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchMemoryUseCase_Execute(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockMemoryRepository
		args       func() SearchMemoryRequest
		assert     func(t *testing.T, results []domain.SearchResult, err error)
	}{
		{
			name: "searches with text only",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Search(mock.Anything, mock.MatchedBy(func(q domain.SearchQuery) bool {
					return q.Text == "auth bug" && q.Limit == 10
				})).Return([]domain.SearchResult{{}, {}}, nil)
				return m
			},
			args: func() SearchMemoryRequest { return SearchMemoryRequest{Text: "auth bug"} },
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 2)
			},
		},
		{
			name: "passes filters to repository",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Search(mock.Anything, mock.MatchedBy(func(q domain.SearchQuery) bool {
					return q.Type == domain.TypeBugfix &&
						q.Project == "myapp" &&
						q.Scope == domain.ScopeProject &&
						q.Field == "location" &&
						q.Limit == 5
				})).Return([]domain.SearchResult{{}}, nil)
				return m
			},
			args: func() SearchMemoryRequest {
				return SearchMemoryRequest{
					Text: "auth", Type: "bugfix", Project: "myapp",
					Scope: "project", Field: "location", Limit: 5,
				}
			},
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				assert.NoError(t, err)
				assert.Len(t, results, 1)
			},
		},
		{
			name:       "rejects empty text",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args:       func() SearchMemoryRequest { return SearchMemoryRequest{Text: ""} },
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptySearchQuery)
			},
		},
		{
			name:       "rejects invalid type",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args:       func() SearchMemoryRequest { return SearchMemoryRequest{Text: "auth", Type: "invalid"} },
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidMemoryType)
			},
		},
		{
			name:       "rejects invalid scope",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args:       func() SearchMemoryRequest { return SearchMemoryRequest{Text: "auth", Scope: "global"} },
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidScope)
			},
		},
		{
			name:       "rejects invalid field",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args:       func() SearchMemoryRequest { return SearchMemoryRequest{Text: "auth", Field: "foobar"} },
			assert: func(t *testing.T, results []domain.SearchResult, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidField)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			req := tc.args()

			uc := NewSearchMemoryUseCase(repo)
			results, err := uc.Execute(context.Background(), req)

			tc.assert(t, results, err)
			repo.AssertExpectations(t)
		})
	}
}
