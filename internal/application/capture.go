package application

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sherlook22/cortex/internal/domain"
)

// CaptureRequest holds the input for the Capture use case.
type CaptureRequest struct {
	Input     string
	Project   string
	SessionID string
	Source    string
}

// CaptureUseCase handles extracting and saving learnings from raw text.
type CaptureUseCase struct {
	repo domain.MemoryRepository
}

// NewCaptureUseCase creates a new CaptureUseCase.
func NewCaptureUseCase(repo domain.MemoryRepository) *CaptureUseCase {
	return &CaptureUseCase{repo: repo}
}

// Execute parses raw text input and saves extracted learnings as memories.
// Returns the number of memories created.
func (uc *CaptureUseCase) Execute(ctx context.Context, req CaptureRequest) (int, error) {
	input := strings.TrimSpace(req.Input)
	if input == "" {
		return 0, domain.ErrEmptyCaptureInput
	}
	if strings.TrimSpace(req.Project) == "" {
		return 0, domain.ErrEmptyProject
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	source := req.Source
	if source == "" {
		source = "manual"
	}

	items := extractLearnings(input)

	if len(items) == 0 {
		// No structured learnings found; save the entire block as one memory.
		title := generateTitle(input)
		memory := &domain.Memory{
			Title:     title,
			Type:      domain.TypeDiscovery,
			Project:   strings.ToLower(strings.TrimSpace(req.Project)),
			Scope:     domain.ScopeProject,
			What:      truncateContent(input, 2000),
			Why:       "Captured from " + source,
			Location:  "",
			Learned:   truncateContent(input, 2000),
			SessionID: sessionID,
			Source:    source,
		}

		if _, err := uc.repo.Save(ctx, memory); err != nil {
			return 0, fmt.Errorf("saving captured memory: %w", err)
		}
		return 1, nil
	}

	// Save each extracted learning as a separate memory.
	count := 0
	for _, item := range items {
		memory := &domain.Memory{
			Title:     generateTitle(item),
			Type:      domain.TypeDiscovery,
			Project:   strings.ToLower(strings.TrimSpace(req.Project)),
			Scope:     domain.ScopeProject,
			What:      truncateContent(item, 2000),
			Why:       "Captured from " + source,
			Location:  "",
			Learned:   truncateContent(item, 2000),
			SessionID: sessionID,
			Source:    source,
		}

		if _, err := uc.repo.Save(ctx, memory); err != nil {
			return count, fmt.Errorf("saving captured learning: %w", err)
		}
		count++
	}

	return count, nil
}

// sectionHeaderRe matches markdown headers like "## Key Learnings:" or "## Discoveries"
var sectionHeaderRe = regexp.MustCompile(`(?i)^##\s+(key\s+learnings?|learnings?|discoveries?|discovered|aprendizajes)\s*:?\s*$`)

// listItemRe matches bulleted or numbered list items.
var listItemRe = regexp.MustCompile(`^\s*(?:[-*+]|\d+[.)]\s)(.+)`)

// extractLearnings attempts to parse structured learnings from text.
// Returns nil if no structured content is found.
func extractLearnings(text string) []string {
	lines := strings.Split(text, "\n")
	var items []string
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if sectionHeaderRe.MatchString(trimmed) {
			inSection = true
			continue
		}

		// If we hit another header while in a section, stop.
		if inSection && strings.HasPrefix(trimmed, "#") {
			break
		}

		if inSection {
			if m := listItemRe.FindStringSubmatch(line); m != nil {
				item := strings.TrimSpace(m[1])
				if item != "" {
					items = append(items, item)
				}
			}
		}
	}

	return items
}

// generateTitle creates a short title from the first words of text.
func generateTitle(text string) string {
	// Take the first line or first 60 chars.
	first := strings.SplitN(text, "\n", 2)[0]
	first = strings.TrimSpace(first)

	// Remove markdown formatting.
	first = strings.TrimLeft(first, "#*-+ ")
	first = strings.TrimSpace(first)

	if len(first) > 60 {
		first = first[:57] + "..."
	}

	if first == "" {
		return "Captured learning"
	}

	return first
}

// truncateContent limits text to maxLen characters.
func truncateContent(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
