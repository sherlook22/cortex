package application

import (
	"context"

	"github.com/sherlook22/cortex/internal/domain"
)

// UpdateMemoryRequest holds the input for the UpdateMemory use case.
// Nil pointers mean "do not update this field".
type UpdateMemoryRequest struct {
	ID       int64
	Title    *string
	Type     *string
	What     *string
	Why      *string
	Location *string
	Learned  *string
	Tags     *[]string
	TopicKey *string
}

// UpdateMemoryUseCase handles modifying an existing memory.
type UpdateMemoryUseCase struct {
	repo domain.MemoryRepository
}

// NewUpdateMemoryUseCase creates a new UpdateMemoryUseCase.
func NewUpdateMemoryUseCase(repo domain.MemoryRepository) *UpdateMemoryUseCase {
	return &UpdateMemoryUseCase{repo: repo}
}

// Execute validates and updates a memory.
func (uc *UpdateMemoryUseCase) Execute(ctx context.Context, req UpdateMemoryRequest) error {
	if req.Type != nil && !domain.IsValidMemoryType(*req.Type) {
		return domain.ErrInvalidMemoryType
	}

	params := domain.UpdateParams{
		Title:    req.Title,
		What:     req.What,
		Why:      req.Why,
		Location: req.Location,
		Learned:  req.Learned,
		Tags:     req.Tags,
		TopicKey: req.TopicKey,
	}

	if req.Type != nil {
		t := domain.MemoryType(*req.Type)
		params.Type = &t
	}

	return uc.repo.Update(ctx, req.ID, params)
}
