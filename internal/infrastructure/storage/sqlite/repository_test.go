package sqlite

import (
	"context"
	"fmt"
	"testing"

	"github.com/sherlook22/cortex/internal/domain"
)

func setupTestRepo(t *testing.T) *Repository {
	t.Helper()
	db, err := OpenInMemory()
	if err != nil {
		t.Fatalf("OpenInMemory() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewRepository(db)
}

func newTestMemory() *domain.Memory {
	return &domain.Memory{
		Title:    "Fixed auth bug",
		Type:     domain.TypeBugfix,
		Project:  "myapp",
		Scope:    domain.ScopeProject,
		What:     "Sanitized user input in query builder",
		Why:      "SQL injection vulnerability found in audit",
		Location: "src/db/query.go:142",
		Learned:  "Always use parameterized queries",
	}
}

func TestRepository_Save(t *testing.T) {
	tests := []struct {
		name    string
		memory  *domain.Memory
		wantErr bool
	}{
		{
			name:    "saves a basic memory",
			memory:  newTestMemory(),
			wantErr: false,
		},
		{
			name: "saves with tags",
			memory: func() *domain.Memory {
				m := newTestMemory()
				m.Tags = []string{"auth", "security"}
				return m
			}(),
			wantErr: false,
		},
		{
			name: "saves with topic key",
			memory: func() *domain.Memory {
				m := newTestMemory()
				m.TopicKey = "architecture/auth-model"
				return m
			}(),
			wantErr: false,
		},
		{
			name: "saves with personal scope",
			memory: func() *domain.Memory {
				m := newTestMemory()
				m.Scope = domain.ScopePersonal
				return m
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestRepo(t)
			ctx := context.Background()

			id, err := repo.Save(ctx, tt.memory)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && id <= 0 {
				t.Errorf("Save() returned id = %d, want > 0", id)
			}
		})
	}
}

func TestRepository_Save_TopicKeyUpsert(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Save initial memory with topic key.
	m1 := newTestMemory()
	m1.TopicKey = "architecture/auth"
	id1, err := repo.Save(ctx, m1)
	if err != nil {
		t.Fatalf("first Save() error = %v", err)
	}

	// Save again with same topic key — should upsert.
	m2 := newTestMemory()
	m2.TopicKey = "architecture/auth"
	m2.Title = "Updated auth decision"
	m2.What = "Switched to JWT"
	id2, err := repo.Save(ctx, m2)
	if err != nil {
		t.Fatalf("second Save() error = %v", err)
	}

	if id1 != id2 {
		t.Errorf("upsert returned different IDs: %d vs %d", id1, id2)
	}

	// Verify the content was updated.
	got, err := repo.GetByID(ctx, id1)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Title != "Updated auth decision" {
		t.Errorf("title = %q, want %q", got.Title, "Updated auth decision")
	}
	if got.What != "Switched to JWT" {
		t.Errorf("what = %q, want %q", got.What, "Switched to JWT")
	}
}

func TestRepository_Save_DifferentTopicKeys(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	m1 := newTestMemory()
	m1.TopicKey = "architecture/auth"
	id1, err := repo.Save(ctx, m1)
	if err != nil {
		t.Fatalf("first Save() error = %v", err)
	}

	m2 := newTestMemory()
	m2.TopicKey = "architecture/db"
	id2, err := repo.Save(ctx, m2)
	if err != nil {
		t.Fatalf("second Save() error = %v", err)
	}

	if id1 == id2 {
		t.Error("different topic keys should create different memories")
	}
}

func TestRepository_GetByID(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() int64
		id      int64
		wantErr error
	}{
		{
			name: "retrieves existing memory",
			setup: func() int64 {
				id, _ := repo.Save(ctx, newTestMemory())
				return id
			},
			wantErr: nil,
		},
		{
			name:    "returns error for non-existent ID",
			setup:   func() int64 { return 9999 },
			wantErr: domain.ErrMemoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.setup()
			got, err := repo.GetByID(ctx, id)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("GetByID() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByID() error = %v", err)
			}
			if got.ID != id {
				t.Errorf("ID = %d, want %d", got.ID, id)
			}
			if got.Title != "Fixed auth bug" {
				t.Errorf("Title = %q, want %q", got.Title, "Fixed auth bug")
			}
		})
	}
}

func TestRepository_Search(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Seed data.
	memories := []*domain.Memory{
		{
			Title: "Fixed auth bug", Type: domain.TypeBugfix, Project: "myapp",
			Scope: domain.ScopeProject, What: "Sanitized input", Why: "SQL injection",
			Location: "src/db/query.go", Learned: "Use parameterized queries",
			Tags: []string{"auth", "security"},
		},
		{
			Title: "Database migration strategy", Type: domain.TypeDecision, Project: "myapp",
			Scope: domain.ScopeProject, What: "Chose goose for migrations", Why: "Simple and reliable",
			Location: "src/db/migrations/", Learned: "Goose supports both SQL and Go migrations",
		},
		{
			Title: "Docker setup", Type: domain.TypeConfig, Project: "other-project",
			Scope: domain.ScopeProject, What: "Added docker-compose", Why: "Local dev environment",
			Location: "docker-compose.yml", Learned: "Use named volumes for persistence",
		},
	}
	for _, m := range memories {
		if _, err := repo.Save(ctx, m); err != nil {
			t.Fatalf("seeding memory: %v", err)
		}
	}

	tests := []struct {
		name      string
		query     domain.SearchQuery
		wantCount int
	}{
		{
			name:      "search by general term",
			query:     domain.SearchQuery{Text: "auth", Limit: 10},
			wantCount: 1,
		},
		{
			name:      "search cross-column (injection is in why, query in location)",
			query:     domain.SearchQuery{Text: "injection", Limit: 10},
			wantCount: 1,
		},
		{
			name:      "search with project filter",
			query:     domain.SearchQuery{Text: "migrations", Project: "myapp", Limit: 10},
			wantCount: 1,
		},
		{
			name:      "search with type filter",
			query:     domain.SearchQuery{Text: "docker", Type: domain.TypeConfig, Limit: 10},
			wantCount: 1,
		},
		{
			name:      "search with no results",
			query:     domain.SearchQuery{Text: "nonexistent", Limit: 10},
			wantCount: 0,
		},
		{
			name:      "search by specific field (tags)",
			query:     domain.SearchQuery{Text: "security", Field: "tags", Limit: 10},
			wantCount: 1,
		},
		{
			name:      "search by specific field (location)",
			query:     domain.SearchQuery{Text: "docker-compose", Field: "location", Limit: 10},
			wantCount: 1,
		},
		{
			name:      "search respects limit",
			query:     domain.SearchQuery{Text: "src", Limit: 1},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.Search(ctx, tt.query)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}
			if len(results) != tt.wantCount {
				t.Errorf("Search() returned %d results, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestRepository_Update(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	id, err := repo.Save(ctx, newTestMemory())
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	newTitle := "Updated title"
	newWhat := "Updated what"
	newType := domain.TypeDecision

	tests := []struct {
		name    string
		id      int64
		params  domain.UpdateParams
		wantErr error
	}{
		{
			name:    "update title",
			id:      id,
			params:  domain.UpdateParams{Title: &newTitle},
			wantErr: nil,
		},
		{
			name:    "update multiple fields",
			id:      id,
			params:  domain.UpdateParams{What: &newWhat, Type: &newType},
			wantErr: nil,
		},
		{
			name:    "update non-existent memory",
			id:      9999,
			params:  domain.UpdateParams{Title: &newTitle},
			wantErr: domain.ErrMemoryNotFound,
		},
		{
			name:    "update with empty params is no-op",
			id:      id,
			params:  domain.UpdateParams{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(ctx, tt.id, tt.params)
			if err != tt.wantErr {
				t.Errorf("Update() error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	// Verify final state.
	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Title != newTitle {
		t.Errorf("Title = %q, want %q", got.Title, newTitle)
	}
	if got.What != newWhat {
		t.Errorf("What = %q, want %q", got.What, newWhat)
	}
	if got.Type != newType {
		t.Errorf("Type = %q, want %q", got.Type, newType)
	}
}

func TestRepository_Delete(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	id, err := repo.Save(ctx, newTestMemory())
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	tests := []struct {
		name    string
		id      int64
		wantErr error
	}{
		{
			name:    "delete existing memory",
			id:      id,
			wantErr: nil,
		},
		{
			name:    "delete non-existent memory",
			id:      9999,
			wantErr: domain.ErrMemoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			if err != tt.wantErr {
				t.Errorf("Delete() error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	// Verify it's actually gone.
	_, err = repo.GetByID(ctx, id)
	if err != domain.ErrMemoryNotFound {
		t.Errorf("memory should be deleted, got error = %v", err)
	}
}

func TestRepository_GetRecent(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Save 3 memories across 2 projects.
	for i, proj := range []string{"myapp", "myapp", "other"} {
		m := newTestMemory()
		m.Title = fmt.Sprintf("Memory %d", i+1)
		m.Project = proj
		if _, err := repo.Save(ctx, m); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	tests := []struct {
		name      string
		project   string
		limit     int
		wantCount int
	}{
		{
			name:      "all recent memories",
			project:   "",
			limit:     10,
			wantCount: 3,
		},
		{
			name:      "filtered by project",
			project:   "myapp",
			limit:     10,
			wantCount: 2,
		},
		{
			name:      "respects limit",
			project:   "",
			limit:     1,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memories, err := repo.GetRecent(ctx, tt.project, tt.limit)
			if err != nil {
				t.Fatalf("GetRecent() error = %v", err)
			}
			if len(memories) != tt.wantCount {
				t.Errorf("GetRecent() returned %d memories, want %d", len(memories), tt.wantCount)
			}
		})
	}
}

func TestRepository_GetStats(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	// Empty stats.
	stats, err := repo.GetStats(ctx, "")
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}
	if stats.TotalMemories != 0 {
		t.Errorf("empty DB: TotalMemories = %d, want 0", stats.TotalMemories)
	}

	// Seed data.
	types := []domain.MemoryType{domain.TypeBugfix, domain.TypeBugfix, domain.TypeDecision}
	projects := []string{"app1", "app1", "app2"}
	for i := range types {
		m := newTestMemory()
		m.Type = types[i]
		m.Project = projects[i]
		if _, err := repo.Save(ctx, m); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	tests := []struct {
		name         string
		project      string
		wantTotal    int
		wantTypes    map[domain.MemoryType]int
		wantProjects map[string]int
	}{
		{
			name:         "global stats",
			project:      "",
			wantTotal:    3,
			wantTypes:    map[domain.MemoryType]int{domain.TypeBugfix: 2, domain.TypeDecision: 1},
			wantProjects: map[string]int{"app1": 2, "app2": 1},
		},
		{
			name:         "filtered by project",
			project:      "app1",
			wantTotal:    2,
			wantTypes:    map[domain.MemoryType]int{domain.TypeBugfix: 2},
			wantProjects: map[string]int{"app1": 2, "app2": 1}, // ByProject is always global
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := repo.GetStats(ctx, tt.project)
			if err != nil {
				t.Fatalf("GetStats() error = %v", err)
			}
			if stats.TotalMemories != tt.wantTotal {
				t.Errorf("TotalMemories = %d, want %d", stats.TotalMemories, tt.wantTotal)
			}
			for typ, want := range tt.wantTypes {
				if got := stats.ByType[typ]; got != want {
					t.Errorf("ByType[%s] = %d, want %d", typ, got, want)
				}
			}
		})
	}
}
