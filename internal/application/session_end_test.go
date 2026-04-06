package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEndSessionUseCase_Execute(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockSessionRepository
		args       EndSessionRequest
		assert     func(t *testing.T, err error)
	}{
		{
			name: "ends session successfully",
			setupMocks: func() *mocks.MockSessionRepository {
				m := mocks.NewMockSessionRepository(t)
				m.EXPECT().EndSession(mock.Anything, "sess-1", "done").Return(nil)
				return m
			},
			args: EndSessionRequest{ID: "sess-1", Summary: "done"},
			assert: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:       "rejects empty ID",
			setupMocks: func() *mocks.MockSessionRepository { return mocks.NewMockSessionRepository(t) },
			args:       EndSessionRequest{ID: ""},
			assert: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptySessionID)
			},
		},
		{
			name: "not found",
			setupMocks: func() *mocks.MockSessionRepository {
				m := mocks.NewMockSessionRepository(t)
				m.EXPECT().EndSession(mock.Anything, "no-exist", "").Return(domain.ErrSessionNotFound)
				return m
			},
			args: EndSessionRequest{ID: "no-exist"},
			assert: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			uc := NewEndSessionUseCase(repo)
			err := uc.Execute(context.Background(), tc.args)
			tc.assert(t, err)
			repo.AssertExpectations(t)
		})
	}
}
