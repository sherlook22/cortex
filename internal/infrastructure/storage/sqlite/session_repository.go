package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sherlook22/cortex/internal/domain"
)

// SessionRepository implements domain.SessionRepository backed by SQLite.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new SQLite-backed session repository.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession persists a new session. Idempotent: if a session with the same
// ID already exists, it returns without error.
func (r *SessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO sessions (id, project, directory, status)
		 VALUES (?, ?, ?, ?)`,
		session.ID,
		session.Project,
		session.Directory,
		string(domain.SessionActive),
	)
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	return nil
}

// EndSession closes a session by setting its status to completed and storing the summary.
func (r *SessionRepository) EndSession(ctx context.Context, id string, summary string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE sessions
		 SET status = ?, summary = ?, updated_at = datetime('now')
		 WHERE id = ?`,
		string(domain.SessionCompleted),
		summary,
		id,
	)
	if err != nil {
		return fmt.Errorf("ending session %s: %w", id, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking affected rows: %w", err)
	}
	if affected == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

// GetSession retrieves a single session by its ID.
func (r *SessionRepository) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, project, directory, status, summary, created_at, updated_at
		 FROM sessions WHERE id = ?`, id)

	s, err := scanSession(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("getting session %s: %w", id, err)
	}
	return s, nil
}

// ListSessions returns recent sessions, optionally filtered by project.
func (r *SessionRepository) ListSessions(ctx context.Context, project string, limit int) ([]domain.Session, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `SELECT id, project, directory, status, summary, created_at, updated_at FROM sessions`
	args := []any{}

	if project != "" {
		query += " WHERE project = ?"
		args = append(args, project)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning session row: %w", err)
		}
		sessions = append(sessions, *s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating session rows: %w", err)
	}

	return sessions, nil
}

// scannable is already defined in repository.go

func scanSession(s scannable) (*domain.Session, error) {
	var sess domain.Session
	var status string
	var createdAt, updatedAt string

	err := s.Scan(
		&sess.ID, &sess.Project, &sess.Directory, &status,
		&sess.Summary, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	sess.Status = domain.SessionStatus(status)
	sess.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	sess.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return &sess, nil
}
