package application

import (
	"context"

	"github.com/sherlook22/cortex/internal/domain"
)

// DeleteMemoryUseCase handles removing a memory.
type DeleteMemoryUseCase struct {
	repo domain.MemoryRepository
}

// NewDeleteMemoryUseCase creates a new DeleteMemoryUseCase.
func NewDeleteMemoryUseCase(repo domain.MemoryRepository) *DeleteMemoryUseCase {
	return &DeleteMemoryUseCase{repo: repo}
}

// Execute permanently deletes a memory by its ID.
func (uc *DeleteMemoryUseCase) Execute(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}
