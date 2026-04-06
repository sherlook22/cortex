package application

import (
	"context"
	"fmt"

	"github.com/sherlook22/cortex/internal/domain"
)

// ListSessionsRequest holds the input for the ListSessions use case.
type ListSessionsRequest struct {
	Project string
	Limit   int
}

// ListSessionsUseCase handles listing recent sessions.
type ListSessionsUseCase struct {
	repo domain.SessionRepository
}

// NewListSessionsUseCase creates a new ListSessionsUseCase.
func NewListSessionsUseCase(repo domain.SessionRepository) *ListSessionsUseCase {
	return &ListSessionsUseCase{repo: repo}
}

// Execute returns recent sessions, optionally filtered by project.
func (uc *ListSessionsUseCase) Execute(ctx context.Context, req ListSessionsRequest) ([]domain.Session, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	sessions, err := uc.repo.ListSessions(ctx, req.Project, limit)
	if err != nil {
		return nil, fmt.Errorf("listing sessions: %w", err)
	}

	return sessions, nil
}
