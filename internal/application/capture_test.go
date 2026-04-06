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

func TestCaptureUseCase_Execute(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockMemoryRepository
		args       CaptureRequest
		assert     func(t *testing.T, count int, err error)
	}{
		{
			name: "saves unstructured text as single memory",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Save(mock.Anything, mock.MatchedBy(func(mem *domain.Memory) bool {
					return mem.Type == domain.TypeDiscovery && mem.Project == "myapp" && mem.Source == "subagent"
				})).Return(int64(1), nil)
				return m
			},
			args: CaptureRequest{Input: "some raw output text", Project: "myapp", Source: "subagent"},
			assert: func(t *testing.T, count int, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
		{
			name: "extracts structured learnings",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Save(mock.Anything, mock.Anything).Return(int64(1), nil).Times(2)
				return m
			},
			args: CaptureRequest{
				Input:   "## Key Learnings:\n- First learning\n- Second learning\n",
				Project: "myapp",
			},
			assert: func(t *testing.T, count int, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 2, count)
			},
		},
		{
			name: "handles numbered learnings",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Save(mock.Anything, mock.Anything).Return(int64(1), nil).Times(3)
				return m
			},
			args: CaptureRequest{
				Input:   "## Discoveries\n1. First discovery\n2. Second discovery\n3. Third discovery",
				Project: "myapp",
			},
			assert: func(t *testing.T, count int, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 3, count)
			},
		},
		{
			name:       "rejects empty input",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args:       CaptureRequest{Input: "", Project: "myapp"},
			assert: func(t *testing.T, count int, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptyCaptureInput)
			},
		},
		{
			name:       "rejects empty project",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args:       CaptureRequest{Input: "some text", Project: ""},
			assert: func(t *testing.T, count int, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptyProject)
			},
		},
		{
			name: "propagates repo error on unstructured",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Save(mock.Anything, mock.Anything).Return(int64(0), errors.New("db error"))
				return m
			},
			args: CaptureRequest{Input: "some text", Project: "myapp"},
			assert: func(t *testing.T, count int, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "db error")
			},
		},
		{
			name: "propagates repo error on structured",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Save(mock.Anything, mock.Anything).Return(int64(0), errors.New("db error"))
				return m
			},
			args: CaptureRequest{Input: "## Key Learnings:\n- Item one\n- Item two", Project: "myapp"},
			assert: func(t *testing.T, count int, err error) {
				assert.Error(t, err)
				assert.Equal(t, 0, count)
			},
		},
		{
			name: "defaults source to manual",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Save(mock.Anything, mock.MatchedBy(func(mem *domain.Memory) bool {
					return mem.Source == "manual"
				})).Return(int64(1), nil)
				return m
			},
			args: CaptureRequest{Input: "some text", Project: "myapp"},
			assert: func(t *testing.T, count int, err error) {
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			uc := NewCaptureUseCase(repo)
			count, err := uc.Execute(context.Background(), tc.args)
			tc.assert(t, count, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestExtractLearnings(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect int
	}{
		{"key learnings with bullets", "## Key Learnings:\n- Item one\n- Item two", 2},
		{"discoveries with numbers", "## Discoveries\n1. First\n2. Second\n3. Third", 3},
		{"case insensitive", "## KEY LEARNINGS\n- Item", 1},
		{"no structure", "Just plain text without any headers", 0},
		{"stops at next header", "## Learnings:\n- Item one\n## Other\n- Ignored", 1},
		{"empty items ignored", "## Learnings:\n- \n- Real item\n-  ", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			items := extractLearnings(tc.input)
			assert.Len(t, items, tc.expect)
		})
	}
}

func TestGenerateTitle(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect string
	}{
		{"short text", "Fixed auth bug", "Fixed auth bug"},
		{"long text truncated", "This is a very long text that exceeds sixty characters and should be truncated properly", "This is a very long text that exceeds sixty characters an..."},
		{"strips markdown", "## Some Header", "Some Header"},
		{"empty text", "", "Captured learning"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := generateTitle(tc.input)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestTruncateContent(t *testing.T) {
	assert.Equal(t, "short", truncateContent("short", 100))
	assert.Equal(t, "ab...", truncateContent("abcdefgh", 5))
}
