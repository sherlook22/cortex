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

func TestStartSessionUseCase_Execute(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockSessionRepository
		args       StartSessionRequest
		assert     func(t *testing.T, err error)
	}{
		{
			name: "creates session successfully",
			setupMocks: func() *mocks.MockSessionRepository {
				m := mocks.NewMockSessionRepository(t)
				m.EXPECT().CreateSession(mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.ID == "sess-1" && s.Project == "myapp" && s.Directory == "/home/dev"
				})).Return(nil)
				return m
			},
			args: StartSessionRequest{ID: "sess-1", Project: "myapp", Directory: "/home/dev"},
			assert: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "normalizes project to lowercase",
			setupMocks: func() *mocks.MockSessionRepository {
				m := mocks.NewMockSessionRepository(t)
				m.EXPECT().CreateSession(mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.Project == "myapp"
				})).Return(nil)
				return m
			},
			args: StartSessionRequest{ID: "sess-1", Project: "MyApp"},
			assert: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:       "rejects empty ID",
			setupMocks: func() *mocks.MockSessionRepository { return mocks.NewMockSessionRepository(t) },
			args:       StartSessionRequest{ID: "", Project: "myapp"},
			assert: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptySessionID)
			},
		},
		{
			name:       "rejects empty project",
			setupMocks: func() *mocks.MockSessionRepository { return mocks.NewMockSessionRepository(t) },
			args:       StartSessionRequest{ID: "sess-1", Project: ""},
			assert: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptyProject)
			},
		},
		{
			name: "propagates repo error",
			setupMocks: func() *mocks.MockSessionRepository {
				m := mocks.NewMockSessionRepository(t)
				m.EXPECT().CreateSession(mock.Anything, mock.Anything).Return(errors.New("db error"))
				return m
			},
			args: StartSessionRequest{ID: "sess-1", Project: "myapp"},
			assert: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "db error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			uc := NewStartSessionUseCase(repo)
			err := uc.Execute(context.Background(), tc.args)
			tc.assert(t, err)
			repo.AssertExpectations(t)
		})
	}
}
