package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/sherlook22/cortex/internal/domain"
)

// GetSessionUseCase handles retrieving a single session.
type GetSessionUseCase struct {
	repo domain.SessionRepository
}

// NewGetSessionUseCase creates a new GetSessionUseCase.
func NewGetSessionUseCase(repo domain.SessionRepository) *GetSessionUseCase {
	return &GetSessionUseCase{repo: repo}
}

// Execute retrieves a session by its ID.
func (uc *GetSessionUseCase) Execute(ctx context.Context, id string) (*domain.Session, error) {
	if strings.TrimSpace(id) == "" {
		return nil, domain.ErrEmptySessionID
	}

	session, err := uc.repo.GetSession(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}

	return session, nil
}
