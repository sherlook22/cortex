package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSaveMemoryUseCase_Execute(t *testing.T) {
	validReq := SaveMemoryRequest{
		Title:    "Fixed auth bug",
		Type:     "bugfix",
		Project:  "myapp",
		What:     "Sanitized user input",
		Why:      "SQL injection vulnerability",
		Location: "src/db/query.go:142",
		Learned:  "Always use parameterized queries",
	}

	testCases := []struct {
		name       string
		setupMocks func() *mocks.MockMemoryRepository
		args       func() SaveMemoryRequest
		assert     func(t *testing.T, id int64, err error)
	}{
		{
			name: "saves valid memory",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.Memory")).Return(int64(1), nil)
				return m
			},
			args: func() SaveMemoryRequest { return validReq },
			assert: func(t *testing.T, id int64, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), id)
			},
		},
		{
			name: "normalizes project to lowercase",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Save(mock.Anything, mock.MatchedBy(func(mem *domain.Memory) bool {
					return mem.Project == "myapp"
				})).Return(int64(1), nil)
				return m
			},
			args: func() SaveMemoryRequest {
				r := validReq
				r.Project = "  MyApp  "
				return r
			},
			assert: func(t *testing.T, id int64, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), id)
			},
		},
		{
			name: "deduplicates tags",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Save(mock.Anything, mock.MatchedBy(func(mem *domain.Memory) bool {
					return len(mem.Tags) == 2 && mem.Tags[0] == "auth" && mem.Tags[1] == "security"
				})).Return(int64(1), nil)
				return m
			},
			args: func() SaveMemoryRequest {
				r := validReq
				r.Tags = []string{"auth", "AUTH", "security", "auth"}
				return r
			},
			assert: func(t *testing.T, id int64, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), id)
			},
		},
		{
			name: "defaults scope to project",
			setupMocks: func() *mocks.MockMemoryRepository {
				m := mocks.NewMockMemoryRepository(t)
				m.EXPECT().Save(mock.Anything, mock.MatchedBy(func(mem *domain.Memory) bool {
					return mem.Scope == domain.ScopeProject
				})).Return(int64(1), nil)
				return m
			},
			args: func() SaveMemoryRequest { return validReq },
			assert: func(t *testing.T, id int64, err error) {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), id)
			},
		},
		{
			name:       "rejects empty title",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args: func() SaveMemoryRequest {
				r := validReq
				r.Title = ""
				return r
			},
			assert: func(t *testing.T, id int64, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptyTitle)
			},
		},
		{
			name:       "rejects empty project",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args: func() SaveMemoryRequest {
				r := validReq
				r.Project = ""
				return r
			},
			assert: func(t *testing.T, id int64, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptyProject)
			},
		},
		{
			name:       "rejects invalid type",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args: func() SaveMemoryRequest {
				r := validReq
				r.Type = "invalid"
				return r
			},
			assert: func(t *testing.T, id int64, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidMemoryType)
			},
		},
		{
			name:       "rejects invalid scope",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args: func() SaveMemoryRequest {
				r := validReq
				r.Scope = "global"
				return r
			},
			assert: func(t *testing.T, id int64, err error) {
				assert.ErrorIs(t, err, domain.ErrInvalidScope)
			},
		},
		{
			name:       "rejects empty what",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args: func() SaveMemoryRequest {
				r := validReq
				r.What = "  "
				return r
			},
			assert: func(t *testing.T, id int64, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptyWhat)
			},
		},
		{
			name:       "rejects empty why",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args: func() SaveMemoryRequest {
				r := validReq
				r.Why = ""
				return r
			},
			assert: func(t *testing.T, id int64, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptyWhy)
			},
		},
		{
			name:       "rejects empty location",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args: func() SaveMemoryRequest {
				r := validReq
				r.Location = ""
				return r
			},
			assert: func(t *testing.T, id int64, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptyLocation)
			},
		},
		{
			name:       "rejects empty learned",
			setupMocks: func() *mocks.MockMemoryRepository { return mocks.NewMockMemoryRepository(t) },
			args: func() SaveMemoryRequest {
				r := validReq
				r.Learned = ""
				return r
			},
			assert: func(t *testing.T, id int64, err error) {
				assert.ErrorIs(t, err, domain.ErrEmptyLearned)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.setupMocks()
			req := tc.args()

			uc := NewSaveMemoryUseCase(repo)
			id, err := uc.Execute(context.Background(), req)

			tc.assert(t, id, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestNormalizeTags(t *testing.T) {
	testCases := []struct {
		name   string
		args   func() []string
		assert func(t *testing.T, result []string)
	}{
		{
			name: "nil input",
			args: func() []string { return nil },
			assert: func(t *testing.T, result []string) {
				assert.Nil(t, result)
			},
		},
		{
			name: "empty input",
			args: func() []string { return []string{} },
			assert: func(t *testing.T, result []string) {
				assert.Nil(t, result)
			},
		},
		{
			name: "normalizes to lowercase",
			args: func() []string { return []string{"Auth", "DB"} },
			assert: func(t *testing.T, result []string) {
				expected := []string{"auth", "db"}
				assert.Equal(t, expected, result)
			},
		},
		{
			name: "removes duplicates",
			args: func() []string { return []string{"auth", "auth", "db"} },
			assert: func(t *testing.T, result []string) {
				expected := []string{"auth", "db"}
				assert.Equal(t, expected, result)
			},
		},
		{
			name: "trims spaces",
			args: func() []string { return []string{" auth ", "  db"} },
			assert: func(t *testing.T, result []string) {
				expected := []string{"auth", "db"}
				assert.Equal(t, expected, result)
			},
		},
		{
			name: "skips empty strings",
			args: func() []string { return []string{"auth", "", "  ", "db"} },
			assert: func(t *testing.T, result []string) {
				expected := []string{"auth", "db"}
				assert.Equal(t, expected, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tags := tc.args()

			result := normalizeTags(tags)

			tc.assert(t, result)
		})
	}
}

func TestNormalizeTopicKey(t *testing.T) {
	testCases := []struct {
		name   string
		args   func() string
		assert func(t *testing.T, result string)
	}{
		{
			name: "lowercase",
			args: func() string { return "Architecture/Auth" },
			assert: func(t *testing.T, result string) {
				assert.Equal(t, "architecture/auth", result)
			},
		},
		{
			name: "collapse spaces",
			args: func() string { return "bug  fix  auth" },
			assert: func(t *testing.T, result string) {
				assert.Equal(t, "bug-fix-auth", result)
			},
		},
		{
			name: "trim",
			args: func() string { return "  auth model  " },
			assert: func(t *testing.T, result string) {
				assert.Equal(t, "auth-model", result)
			},
		},
		{
			name: "empty",
			args: func() string { return "" },
			assert: func(t *testing.T, result string) {
				assert.Empty(t, result)
			},
		},
		{
			name: "truncate at 120",
			args: func() string { return string(make([]byte, 200)) },
			assert: func(t *testing.T, result string) {
				assert.Len(t, result, 120)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.args()

			result := normalizeTopicKey(input)

			tc.assert(t, result)
		})
	}
}
