package application

import (
	"context"
	"errors"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListSessionsUseCase_Execute(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockSessionRepository
		args       ListSessionsRequest
		assert     func(t *testing.T, sessions []domain.Session, err error)
	}{
		{
			name: "returns sessions",
			setupMocks: func() *mocks.MockSessionRepository {
				m := mocks.NewMockSessionRepository(t)
				m.EXPECT().ListSessions(mock.Anything, "myapp", 10).Return([]domain.Session{
					{ID: "s1", Project: "myapp", Status: domain.SessionActive},
				}, nil)
				return m
			},
			args: ListSessionsRequest{Project: "myapp"},
			assert: func(t *testing.T, sessions []domain.Session, err error) {
				assert.NoError(t, err)
				assert.Len(t, sessions, 1)
			},
		},
		{
			name: "defaults limit to 10",
			setupMocks: func() *mocks.MockSessionRepository {
				m := mocks.NewMockSessionRepository(t)
				m.EXPECT().ListSessions(mock.Anything, "", 10).Return(nil, nil)
				return m
			},
			args: ListSessionsRequest{},
			assert: func(t *testing.T, sessions []domain.Session, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "propagates repo error",
			setupMocks: func() *mocks.MockSessionRepository {
				m := mocks.NewMockSessionRepository(t)
				m.EXPECT().ListSessions(mock.Anything, "", 10).Return(nil, errors.New("db error"))
				return m
			},
			args: ListSessionsRequest{},
			assert: func(t *testing.T, sessions []domain.Session, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "db error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			uc := NewListSessionsUseCase(repo)
			sessions, err := uc.Execute(context.Background(), tc.args)
			tc.assert(t, sessions, err)
			repo.AssertExpectations(t)
		})
	}
}
