package application

import (
	"context"

	"github.com/sherlook22/cortex/internal/domain"
)

// GetMemoryUseCase handles retrieving a single memory by ID.
type GetMemoryUseCase struct {
	repo domain.MemoryRepository
}

// NewGetMemoryUseCase creates a new GetMemoryUseCase.
func NewGetMemoryUseCase(repo domain.MemoryRepository) *GetMemoryUseCase {
	return &GetMemoryUseCase{repo: repo}
}

// Execute retrieves a memory by its ID.
func (uc *GetMemoryUseCase) Execute(ctx context.Context, id int64) (*domain.Memory, error) {
	return uc.repo.GetByID(ctx, id)
}
