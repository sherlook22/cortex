package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/sherlook22/cortex/internal/domain"
)

// SaveMemoryRequest holds the input for the SaveMemory use case.
type SaveMemoryRequest struct {
	Title    string
	Type     string
	Project  string
	Scope    string
	What     string
	Why      string
	Location string
	Learned  string
	Tags     []string
	TopicKey string
}

// SaveMemoryUseCase handles persisting a new memory.
type SaveMemoryUseCase struct {
	repo domain.MemoryRepository
}

// NewSaveMemoryUseCase creates a new SaveMemoryUseCase.
func NewSaveMemoryUseCase(repo domain.MemoryRepository) *SaveMemoryUseCase {
	return &SaveMemoryUseCase{repo: repo}
}

// Execute validates and saves a memory. Returns the saved memory ID.
func (uc *SaveMemoryUseCase) Execute(ctx context.Context, req SaveMemoryRequest) (int64, error) {
	if err := validateSaveRequest(req); err != nil {
		return 0, err
	}

	scope := domain.Scope(req.Scope)
	if scope == "" {
		scope = domain.ScopeProject
	}

	memory := &domain.Memory{
		Title:    strings.TrimSpace(req.Title),
		Type:     domain.MemoryType(req.Type),
		Project:  strings.ToLower(strings.TrimSpace(req.Project)),
		Scope:    scope,
		What:     strings.TrimSpace(req.What),
		Why:      strings.TrimSpace(req.Why),
		Location: strings.TrimSpace(req.Location),
		Learned:  strings.TrimSpace(req.Learned),
		Tags:     normalizeTags(req.Tags),
		TopicKey: normalizeTopicKey(req.TopicKey),
	}

	id, err := uc.repo.Save(ctx, memory)
	if err != nil {
		return 0, fmt.Errorf("saving memory: %w", err)
	}

	return id, nil
}

func validateSaveRequest(req SaveMemoryRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return domain.ErrEmptyTitle
	}
	if !domain.IsValidMemoryType(req.Type) {
		return domain.ErrInvalidMemoryType
	}
	if strings.TrimSpace(req.Project) == "" {
		return domain.ErrEmptyProject
	}
	if req.Scope != "" && !domain.IsValidScope(req.Scope) {
		return domain.ErrInvalidScope
	}
	if strings.TrimSpace(req.What) == "" {
		return domain.ErrEmptyWhat
	}
	if strings.TrimSpace(req.Why) == "" {
		return domain.ErrEmptyWhy
	}
	if strings.TrimSpace(req.Location) == "" {
		return domain.ErrEmptyLocation
	}
	if strings.TrimSpace(req.Learned) == "" {
		return domain.ErrEmptyLearned
	}
	return nil
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(tags))
	seen := make(map[string]bool)
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t != "" && !seen[t] {
			normalized = append(normalized, t)
			seen[t] = true
		}
	}
	return normalized
}

func normalizeTopicKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	key = strings.Join(strings.Fields(key), "-")
	if len(key) > 120 {
		key = key[:120]
	}
	return key
}
