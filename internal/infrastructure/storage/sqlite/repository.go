package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sherlook22/cortex/internal/domain"
)

// Repository implements domain.MemoryRepository backed by SQLite.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new SQLite-backed repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Save persists a new memory or upserts if a topic_key match exists.
func (r *Repository) Save(ctx context.Context, memory *domain.Memory) (int64, error) {
	// Topic-key upsert: if topic_key is set, look for existing memory.
	if memory.TopicKey != "" {
		existing, err := r.FindByTopicKey(ctx, memory.TopicKey, memory.Project, memory.Scope)
		if err != nil && err != domain.ErrMemoryNotFound {
			return 0, fmt.Errorf("checking topic key: %w", err)
		}
		if existing != nil {
			err := r.upsertByTopicKey(ctx, existing.ID, memory)
			if err != nil {
				return 0, fmt.Errorf("upserting by topic key: %w", err)
			}
			return existing.ID, nil
		}
	}

	tags := joinTags(memory.Tags)

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO memories (title, type, project, scope, what, why, location, learned, tags, topic_key)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		memory.Title,
		string(memory.Type),
		memory.Project,
		string(memory.Scope),
		memory.What,
		memory.Why,
		memory.Location,
		memory.Learned,
		tags,
		nullString(memory.TopicKey),
	)
	if err != nil {
		return 0, fmt.Errorf("inserting memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}

	return id, nil
}

// upsertByTopicKey updates an existing memory matched by topic key.
func (r *Repository) upsertByTopicKey(ctx context.Context, id int64, memory *domain.Memory) error {
	tags := joinTags(memory.Tags)

	_, err := r.db.ExecContext(ctx,
		`UPDATE memories
		 SET title = ?, type = ?, what = ?, why = ?, location = ?, learned = ?,
		     tags = ?, topic_key = ?, updated_at = datetime('now')
		 WHERE id = ?`,
		memory.Title,
		string(memory.Type),
		memory.What,
		memory.Why,
		memory.Location,
		memory.Learned,
		tags,
		nullString(memory.TopicKey),
		id,
	)
	if err != nil {
		return fmt.Errorf("updating memory %d: %w", id, err)
	}
	return nil
}

// Search performs a full-text search across memories.
func (r *Repository) Search(ctx context.Context, query domain.SearchQuery) ([]domain.SearchResult, error) {
	sanitized := sanitizeFTS(query.Text)
	if sanitized == "" {
		return nil, nil
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 10
	}

	var matchClause string
	if query.Field != "" {
		matchClause = fmt.Sprintf("{%s} : %s", query.Field, sanitized)
	} else {
		matchClause = sanitized
	}

	baseQuery := `
		SELECT m.id, m.title, m.type, m.project, m.scope,
		       m.what, m.why, m.location, m.learned,
		       m.tags, m.topic_key, m.created_at, m.updated_at,
		       f.rank
		FROM memories_fts f
		JOIN memories m ON m.id = f.rowid
		WHERE memories_fts MATCH ?`

	args := []any{matchClause}

	if query.Type != "" {
		baseQuery += " AND m.type = ?"
		args = append(args, string(query.Type))
	}
	if query.Project != "" {
		baseQuery += " AND m.project = ?"
		args = append(args, query.Project)
	}
	if query.Scope != "" {
		baseQuery += " AND m.scope = ?"
		args = append(args, string(query.Scope))
	}

	baseQuery += " ORDER BY f.rank LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("searching memories: %w", err)
	}
	defer rows.Close()

	return scanSearchResults(rows)
}

// GetByID retrieves a single memory by its ID.
func (r *Repository) GetByID(ctx context.Context, id int64) (*domain.Memory, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, title, type, project, scope, what, why, location, learned,
		        tags, topic_key, created_at, updated_at
		 FROM memories WHERE id = ?`, id)

	m, err := scanMemory(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrMemoryNotFound
		}
		return nil, fmt.Errorf("getting memory %d: %w", id, err)
	}
	return m, nil
}

// Update modifies an existing memory. Only non-nil fields are updated.
func (r *Repository) Update(ctx context.Context, id int64, params domain.UpdateParams) error {
	sets := []string{}
	args := []any{}

	if params.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *params.Title)
	}
	if params.Type != nil {
		sets = append(sets, "type = ?")
		args = append(args, string(*params.Type))
	}
	if params.What != nil {
		sets = append(sets, "what = ?")
		args = append(args, *params.What)
	}
	if params.Why != nil {
		sets = append(sets, "why = ?")
		args = append(args, *params.Why)
	}
	if params.Location != nil {
		sets = append(sets, "location = ?")
		args = append(args, *params.Location)
	}
	if params.Learned != nil {
		sets = append(sets, "learned = ?")
		args = append(args, *params.Learned)
	}
	if params.Tags != nil {
		sets = append(sets, "tags = ?")
		args = append(args, joinTags(*params.Tags))
	}
	if params.TopicKey != nil {
		sets = append(sets, "topic_key = ?")
		args = append(args, nullString(*params.TopicKey))
	}

	if len(sets) == 0 {
		return nil
	}

	sets = append(sets, "updated_at = datetime('now')")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE memories SET %s WHERE id = ?", strings.Join(sets, ", "))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating memory %d: %w", id, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking affected rows: %w", err)
	}
	if affected == 0 {
		return domain.ErrMemoryNotFound
	}

	return nil
}

// Delete permanently removes a memory by its ID.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM memories WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting memory %d: %w", id, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking affected rows: %w", err)
	}
	if affected == 0 {
		return domain.ErrMemoryNotFound
	}

	return nil
}

// GetRecent returns the most recent memories, optionally filtered by project.
func (r *Repository) GetRecent(ctx context.Context, project string, limit int) ([]domain.Memory, error) {
	if limit <= 0 {
		limit = 20
	}

	query := "SELECT id, title, type, project, scope, what, why, location, learned, tags, topic_key, created_at, updated_at FROM memories"
	args := []any{}

	if project != "" {
		query += " WHERE project = ?"
		args = append(args, project)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("getting recent memories: %w", err)
	}
	defer rows.Close()

	return scanMemories(rows)
}

// GetStats returns aggregate statistics about stored memories.
func (r *Repository) GetStats(ctx context.Context, project string) (*domain.Stats, error) {
	stats := &domain.Stats{
		ByType:    make(map[domain.MemoryType]int),
		ByProject: make(map[string]int),
	}

	// Total count.
	countQuery := "SELECT COUNT(*) FROM memories"
	countArgs := []any{}
	if project != "" {
		countQuery += " WHERE project = ?"
		countArgs = append(countArgs, project)
	}
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&stats.TotalMemories); err != nil {
		return nil, fmt.Errorf("counting memories: %w", err)
	}

	if stats.TotalMemories == 0 {
		return stats, nil
	}

	// By type.
	typeQuery := "SELECT type, COUNT(*) FROM memories"
	typeArgs := []any{}
	if project != "" {
		typeQuery += " WHERE project = ?"
		typeArgs = append(typeArgs, project)
	}
	typeQuery += " GROUP BY type"

	typeRows, err := r.db.QueryContext(ctx, typeQuery, typeArgs...)
	if err != nil {
		return nil, fmt.Errorf("counting by type: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var t string
		var count int
		if err := typeRows.Scan(&t, &count); err != nil {
			return nil, fmt.Errorf("scanning type count: %w", err)
		}
		stats.ByType[domain.MemoryType(t)] = count
	}

	// By project.
	projRows, err := r.db.QueryContext(ctx, "SELECT project, COUNT(*) FROM memories GROUP BY project")
	if err != nil {
		return nil, fmt.Errorf("counting by project: %w", err)
	}
	defer projRows.Close()

	for projRows.Next() {
		var p string
		var count int
		if err := projRows.Scan(&p, &count); err != nil {
			return nil, fmt.Errorf("scanning project count: %w", err)
		}
		stats.ByProject[p] = count
	}

	// Date range.
	rangeQuery := "SELECT MIN(created_at), MAX(created_at) FROM memories"
	rangeArgs := []any{}
	if project != "" {
		rangeQuery += " WHERE project = ?"
		rangeArgs = append(rangeArgs, project)
	}

	var oldest, newest string
	if err := r.db.QueryRowContext(ctx, rangeQuery, rangeArgs...).Scan(&oldest, &newest); err != nil {
		return nil, fmt.Errorf("getting date range: %w", err)
	}

	stats.OldestMemory, _ = time.Parse("2006-01-02 15:04:05", oldest)
	stats.NewestMemory, _ = time.Parse("2006-01-02 15:04:05", newest)

	return stats, nil
}

// FindByTopicKey finds a memory by its topic key within a project and scope.
func (r *Repository) FindByTopicKey(ctx context.Context, topicKey string, project string, scope domain.Scope) (*domain.Memory, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, title, type, project, scope, what, why, location, learned,
		        tags, topic_key, created_at, updated_at
		 FROM memories
		 WHERE topic_key = ? AND project = ? AND scope = ?
		 ORDER BY datetime(updated_at) DESC
		 LIMIT 1`,
		topicKey, project, string(scope))

	m, err := scanMemory(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrMemoryNotFound
		}
		return nil, fmt.Errorf("finding by topic key: %w", err)
	}
	return m, nil
}

// GetAll returns all memories, optionally filtered by project.
func (r *Repository) GetAll(ctx context.Context, project string) ([]domain.Memory, error) {
	query := "SELECT id, title, type, project, scope, what, why, location, learned, tags, topic_key, created_at, updated_at FROM memories"
	args := []any{}

	if project != "" {
		query += " WHERE project = ?"
		args = append(args, project)
	}

	query += " ORDER BY created_at ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("getting all memories: %w", err)
	}
	defer rows.Close()

	return scanMemories(rows)
}

// SaveImport persists an imported memory, preserving original timestamps.
func (r *Repository) SaveImport(ctx context.Context, memory *domain.Memory) (int64, error) {
	tags := joinTags(memory.Tags)

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO memories (title, type, project, scope, what, why, location, learned, tags, topic_key, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		memory.Title,
		string(memory.Type),
		memory.Project,
		string(memory.Scope),
		memory.What,
		memory.Why,
		memory.Location,
		memory.Learned,
		tags,
		nullString(memory.TopicKey),
		memory.CreatedAt.Format("2006-01-02 15:04:05"),
		memory.UpdatedAt.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return 0, fmt.Errorf("importing memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}

	return id, nil
}

// --- Helpers ---

// scannable abstracts sql.Row and sql.Rows for reusable scanning.
type scannable interface {
	Scan(dest ...any) error
}

func scanMemory(s scannable) (*domain.Memory, error) {
	var m domain.Memory
	var memType, scope, tags, topicKey sql.NullString
	var createdAt, updatedAt string

	err := s.Scan(
		&m.ID, &m.Title, &memType, &m.Project, &scope,
		&m.What, &m.Why, &m.Location, &m.Learned,
		&tags, &topicKey, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	m.Type = domain.MemoryType(memType.String)
	m.Scope = domain.Scope(scope.String)
	m.Tags = splitTags(tags.String)
	m.TopicKey = topicKey.String
	m.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	m.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return &m, nil
}

func scanMemories(rows *sql.Rows) ([]domain.Memory, error) {
	var memories []domain.Memory
	for rows.Next() {
		m, err := scanMemory(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning memory row: %w", err)
		}
		memories = append(memories, *m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}
	return memories, nil
}

func scanSearchResults(rows *sql.Rows) ([]domain.SearchResult, error) {
	var results []domain.SearchResult
	for rows.Next() {
		var m domain.Memory
		var memType, scope, tags, topicKey sql.NullString
		var createdAt, updatedAt string
		var rank float64

		err := rows.Scan(
			&m.ID, &m.Title, &memType, &m.Project, &scope,
			&m.What, &m.Why, &m.Location, &m.Learned,
			&tags, &topicKey, &createdAt, &updatedAt,
			&rank,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning search result: %w", err)
		}

		m.Type = domain.MemoryType(memType.String)
		m.Scope = domain.Scope(scope.String)
		m.Tags = splitTags(tags.String)
		m.TopicKey = topicKey.String
		m.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		m.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		results = append(results, domain.SearchResult{Memory: m, Rank: rank})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating search rows: %w", err)
	}
	return results, nil
}

func joinTags(tags []string) sql.NullString {
	if len(tags) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: strings.Join(tags, ","), Valid: true}
}

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
