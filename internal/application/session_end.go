package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/sherlook22/cortex/internal/domain"
)

// EndSessionRequest holds the input for the EndSession use case.
type EndSessionRequest struct {
	ID      string
	Summary string
}

// EndSessionUseCase handles closing a session with a summary.
type EndSessionUseCase struct {
	repo domain.SessionRepository
}

// NewEndSessionUseCase creates a new EndSessionUseCase.
func NewEndSessionUseCase(repo domain.SessionRepository) *EndSessionUseCase {
	return &EndSessionUseCase{repo: repo}
}

// Execute closes a session and stores its summary.
func (uc *EndSessionUseCase) Execute(ctx context.Context, req EndSessionRequest) error {
	if strings.TrimSpace(req.ID) == "" {
		return domain.ErrEmptySessionID
	}

	if err := uc.repo.EndSession(ctx, req.ID, strings.TrimSpace(req.Summary)); err != nil {
		return fmt.Errorf("ending session: %w", err)
	}

	return nil
}
