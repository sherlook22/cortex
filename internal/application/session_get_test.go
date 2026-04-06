package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSessionUseCase_Execute(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockSessionRepository
		id         string
		assert     func(t *testing.T, session *domain.Session, err error)
	}{
		{
			name: "returns session",
			setupMocks: func() *mocks.MockSessionRepository {
				m := mocks.NewMockSessionRepository(t)
				m.EXPECT().GetSession(mock.Anything, "sess-1").Return(&domain.Session{
					ID: "sess-1", Project: "myapp", Status: domain.SessionActive,
				}, nil)
				return m
			},
			id: "sess-1",
			assert: func(t *testing.T, session *domain.Session, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "sess-1", session.ID)
			},
		},
		{
			name:       "rejects empty ID",
			setupMocks: func() *mocks.MockSessionRepository { return mocks.NewMockSessionRepository(t) },
			id:         "",
			assert: func(t *testing.T, session *domain.Session, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptySessionID)
			},
		},
		{
			name: "not found",
			setupMocks: func() *mocks.MockSessionRepository {
				m := mocks.NewMockSessionRepository(t)
				m.EXPECT().GetSession(mock.Anything, "no-exist").Return(nil, domain.ErrSessionNotFound)
				return m
			},
			id: "no-exist",
			assert: func(t *testing.T, session *domain.Session, err error) {
				assert.Error(t, err)
				assert.Nil(t, session)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			uc := NewGetSessionUseCase(repo)
			session, err := uc.Execute(context.Background(), tc.id)
			tc.assert(t, session, err)
			repo.AssertExpectations(t)
		})
	}
}
