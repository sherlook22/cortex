package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/sherlook22/cortex/internal/domain"
)

// StartSessionRequest holds the input for the StartSession use case.
type StartSessionRequest struct {
	ID        string
	Project   string
	Directory string
}

// StartSessionUseCase handles creating or reopening a session.
type StartSessionUseCase struct {
	repo domain.SessionRepository
}

// NewStartSessionUseCase creates a new StartSessionUseCase.
func NewStartSessionUseCase(repo domain.SessionRepository) *StartSessionUseCase {
	return &StartSessionUseCase{repo: repo}
}

// Execute validates and creates a session. Idempotent: if the session already
// exists, it returns without error.
func (uc *StartSessionUseCase) Execute(ctx context.Context, req StartSessionRequest) error {
	if strings.TrimSpace(req.ID) == "" {
		return domain.ErrEmptySessionID
	}
	if strings.TrimSpace(req.Project) == "" {
		return domain.ErrEmptyProject
	}

	session := &domain.Session{
		ID:        strings.TrimSpace(req.ID),
		Project:   strings.ToLower(strings.TrimSpace(req.Project)),
		Directory: strings.TrimSpace(req.Directory),
		Status:    domain.SessionActive,
	}

	if err := uc.repo.CreateSession(ctx, session); err != nil {
		return fmt.Errorf("starting session: %w", err)
	}

	return nil
}
