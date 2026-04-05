package application

import (
	"context"
	"fmt"

	"github.com/sherlook22/cortex/internal/domain"
)

// SearchMemoryRequest holds the input for the SearchMemory use case.
type SearchMemoryRequest struct {
	Text    string
	Type    string
	Project string
	Scope   string
	Field   string
	Limit   int
}

// SearchMemoryUseCase handles searching memories.
type SearchMemoryUseCase struct {
	repo domain.MemoryRepository
}

// NewSearchMemoryUseCase creates a new SearchMemoryUseCase.
func NewSearchMemoryUseCase(repo domain.MemoryRepository) *SearchMemoryUseCase {
	return &SearchMemoryUseCase{repo: repo}
}

// Execute validates and searches memories. Returns ranked results.
func (uc *SearchMemoryUseCase) Execute(ctx context.Context, req SearchMemoryRequest) ([]domain.SearchResult, error) {
	if req.Text == "" {
		return nil, domain.ErrEmptySearchQuery
	}

	if req.Type != "" && !domain.IsValidMemoryType(req.Type) {
		return nil, domain.ErrInvalidMemoryType
	}
	if req.Scope != "" && !domain.IsValidScope(req.Scope) {
		return nil, domain.ErrInvalidScope
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	query := domain.SearchQuery{
		Text:    req.Text,
		Type:    domain.MemoryType(req.Type),
		Project: req.Project,
		Scope:   domain.Scope(req.Scope),
		Field:   req.Field,
		Limit:   limit,
	}

	results, err := uc.repo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("searching memories: %w", err)
	}

	return results, nil
}
