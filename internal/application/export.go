package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sherlook22/cortex/internal/domain"
)

// ExportUseCase handles exporting memories to JSON.
type ExportUseCase struct {
	repo domain.MemoryRepository
}

// NewExportUseCase creates a new ExportUseCase.
func NewExportUseCase(repo domain.MemoryRepository) *ExportUseCase {
	return &ExportUseCase{repo: repo}
}

// Execute exports memories as a JSON byte slice.
func (uc *ExportUseCase) Execute(ctx context.Context, project string) ([]byte, error) {
	memories, err := uc.repo.GetAll(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("getting memories for export: %w", err)
	}

	data, err := json.MarshalIndent(memories, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling memories: %w", err)
	}

	return data, nil
}
