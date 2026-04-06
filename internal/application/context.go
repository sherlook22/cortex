package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/sherlook22/cortex/internal/domain"
)

// GetContextRequest holds the input for the GetContext use case.
type GetContextRequest struct {
	Project   string
	SessionID string
	Limit     int
}

// GetContextUseCase retrieves recent memories formatted as context.
type GetContextUseCase struct {
	repo domain.MemoryRepository
}

// NewGetContextUseCase creates a new GetContextUseCase.
func NewGetContextUseCase(repo domain.MemoryRepository) *GetContextUseCase {
	return &GetContextUseCase{repo: repo}
}

// Execute retrieves recent memories and formats them as readable context.
func (uc *GetContextUseCase) Execute(ctx context.Context, req GetContextRequest) (string, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	memories, err := uc.repo.GetRecent(ctx, req.Project, req.SessionID, limit)
	if err != nil {
		return "", fmt.Errorf("getting recent memories: %w", err)
	}

	if len(memories) == 0 {
		return "", nil
	}

	return formatContext(memories, req.Project), nil
}

func formatContext(memories []domain.Memory, project string) string {
	var sb strings.Builder

	header := "Recent Memories"
	if project != "" {
		header = fmt.Sprintf("Recent Memories [%s]", project)
	}
	sb.WriteString(fmt.Sprintf("## %s\n\n", header))

	for _, m := range memories {
		sb.WriteString(fmt.Sprintf("### [%d] %s\n", m.ID, m.Title))
		sb.WriteString(fmt.Sprintf("- **Type**: %s | **Project**: %s | **Date**: %s\n",
			m.Type, m.Project, m.CreatedAt.Format("2006-01-02")))
		sb.WriteString(fmt.Sprintf("- **What**: %s\n", m.What))
		sb.WriteString(fmt.Sprintf("- **Why**: %s\n", m.Why))
		sb.WriteString(fmt.Sprintf("- **Location**: %s\n", m.Location))
		sb.WriteString(fmt.Sprintf("- **Learned**: %s\n", m.Learned))
		if len(m.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("- **Tags**: %s\n", strings.Join(m.Tags, ", ")))
		}
		if m.SessionID != "" {
			sb.WriteString(fmt.Sprintf("- **Session**: %s\n", m.SessionID))
		}
		if m.Source != "" {
			sb.WriteString(fmt.Sprintf("- **Source**: %s\n", m.Source))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
