package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sherlook22/cortex/internal/domain"
)

// ImportUseCase handles importing memories from JSON.
type ImportUseCase struct {
	repo domain.MemoryRepository
}

// NewImportUseCase creates a new ImportUseCase.
func NewImportUseCase(repo domain.MemoryRepository) *ImportUseCase {
	return &ImportUseCase{repo: repo}
}

// Execute imports memories from JSON data. Returns the number of imported memories.
func (uc *ImportUseCase) Execute(ctx context.Context, data []byte) (int, error) {
	var memories []domain.Memory
	if err := json.Unmarshal(data, &memories); err != nil {
		return 0, fmt.Errorf("parsing import data: %w", err)
	}

	imported := 0
	for i := range memories {
		if _, err := uc.repo.SaveImport(ctx, &memories[i]); err != nil {
			return imported, fmt.Errorf("importing memory %d (%s): %w", i+1, memories[i].Title, err)
		}
		imported++
	}

	return imported, nil
}
