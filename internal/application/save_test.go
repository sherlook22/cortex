package application

import (
	"context"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
	"github.com/sherlook22/cortex/internal/domain/mocks"
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

	tests := []struct {
		name      string
		req       SaveMemoryRequest
		mockSetup func(*mocks.MockMemoryRepository)
		wantID    int64
		wantErr   error
	}{
		{
			name: "saves valid memory",
			req:  validReq,
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.Memory")).Return(int64(1), nil)
			},
			wantID:  1,
			wantErr: nil,
		},
		{
			name: "normalizes project to lowercase",
			req: func() SaveMemoryRequest {
				r := validReq
				r.Project = "  MyApp  "
				return r
			}(),
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Save(mock.Anything, mock.MatchedBy(func(mem *domain.Memory) bool {
					return mem.Project == "myapp"
				})).Return(int64(1), nil)
			},
			wantID:  1,
			wantErr: nil,
		},
		{
			name: "deduplicates tags",
			req: func() SaveMemoryRequest {
				r := validReq
				r.Tags = []string{"auth", "AUTH", "security", "auth"}
				return r
			}(),
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Save(mock.Anything, mock.MatchedBy(func(mem *domain.Memory) bool {
					return len(mem.Tags) == 2 && mem.Tags[0] == "auth" && mem.Tags[1] == "security"
				})).Return(int64(1), nil)
			},
			wantID:  1,
			wantErr: nil,
		},
		{
			name: "defaults scope to project",
			req:  validReq,
			mockSetup: func(m *mocks.MockMemoryRepository) {
				m.EXPECT().Save(mock.Anything, mock.MatchedBy(func(mem *domain.Memory) bool {
					return mem.Scope == domain.ScopeProject
				})).Return(int64(1), nil)
			},
			wantID:  1,
			wantErr: nil,
		},
		{
			name: "rejects empty title",
			req: func() SaveMemoryRequest {
				r := validReq
				r.Title = ""
				return r
			}(),
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrEmptyTitle,
		},
		{
			name: "rejects empty project",
			req: func() SaveMemoryRequest {
				r := validReq
				r.Project = ""
				return r
			}(),
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrEmptyProject,
		},
		{
			name: "rejects invalid type",
			req: func() SaveMemoryRequest {
				r := validReq
				r.Type = "invalid"
				return r
			}(),
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrInvalidMemoryType,
		},
		{
			name: "rejects invalid scope",
			req: func() SaveMemoryRequest {
				r := validReq
				r.Scope = "global"
				return r
			}(),
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrInvalidScope,
		},
		{
			name: "rejects empty what",
			req: func() SaveMemoryRequest {
				r := validReq
				r.What = "  "
				return r
			}(),
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrEmptyWhat,
		},
		{
			name: "rejects empty why",
			req: func() SaveMemoryRequest {
				r := validReq
				r.Why = ""
				return r
			}(),
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrEmptyWhy,
		},
		{
			name: "rejects empty location",
			req: func() SaveMemoryRequest {
				r := validReq
				r.Location = ""
				return r
			}(),
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrEmptyLocation,
		},
		{
			name: "rejects empty learned",
			req: func() SaveMemoryRequest {
				r := validReq
				r.Learned = ""
				return r
			}(),
			mockSetup: func(m *mocks.MockMemoryRepository) {},
			wantErr:   domain.ErrEmptyLearned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMemoryRepository(t)
			tt.mockSetup(repo)

			uc := NewSaveMemoryUseCase(repo)
			id, err := uc.Execute(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want []string
	}{
		{name: "nil input", tags: nil, want: nil},
		{name: "empty input", tags: []string{}, want: nil},
		{name: "normalizes to lowercase", tags: []string{"Auth", "DB"}, want: []string{"auth", "db"}},
		{name: "removes duplicates", tags: []string{"auth", "auth", "db"}, want: []string{"auth", "db"}},
		{name: "trims spaces", tags: []string{" auth ", "  db"}, want: []string{"auth", "db"}},
		{name: "skips empty strings", tags: []string{"auth", "", "  ", "db"}, want: []string{"auth", "db"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTags(tt.tags)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeTopicKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "lowercase", input: "Architecture/Auth", want: "architecture/auth"},
		{name: "collapse spaces", input: "bug  fix  auth", want: "bug-fix-auth"},
		{name: "trim", input: "  auth model  ", want: "auth-model"},
		{name: "empty", input: "", want: ""},
		{name: "truncate at 120", input: string(make([]byte, 200)), want: string(make([]byte, 120))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTopicKey(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
