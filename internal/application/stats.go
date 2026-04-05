package application

import (
	"context"

	"github.com/sherlook22/cortex/internal/domain"
)

// GetStatsUseCase handles retrieving memory statistics.
type GetStatsUseCase struct {
	repo domain.MemoryRepository
}

// NewGetStatsUseCase creates a new GetStatsUseCase.
func NewGetStatsUseCase(repo domain.MemoryRepository) *GetStatsUseCase {
	return &GetStatsUseCase{repo: repo}
}

// Execute retrieves aggregate statistics, optionally filtered by project.
func (uc *GetStatsUseCase) Execute(ctx context.Context, project string) (*domain.Stats, error) {
	return uc.repo.GetStats(ctx, project)
}
