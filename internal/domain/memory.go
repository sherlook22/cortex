package domain

import "time"

// MemoryType categorizes the kind of knowledge being stored.
type MemoryType string

const (
	TypeBugfix       MemoryType = "bugfix"
	TypeDecision     MemoryType = "decision"
	TypeArchitecture MemoryType = "architecture"
	TypeDiscovery    MemoryType = "discovery"
	TypePattern      MemoryType = "pattern"
	TypeConfig       MemoryType = "config"
)

// ValidMemoryTypes contains all allowed memory types.
var ValidMemoryTypes = []MemoryType{
	TypeBugfix,
	TypeDecision,
	TypeArchitecture,
	TypeDiscovery,
	TypePattern,
	TypeConfig,
}

// IsValidMemoryType checks whether a string is a valid MemoryType.
func IsValidMemoryType(t string) bool {
	for _, valid := range ValidMemoryTypes {
		if MemoryType(t) == valid {
			return true
		}
	}
	return false
}

// Scope defines the visibility of a memory.
type Scope string

const (
	ScopeProject  Scope = "project"
	ScopePersonal Scope = "personal"
)

// ValidScopes contains all allowed scopes.
var ValidScopes = []Scope{ScopeProject, ScopePersonal}

// IsValidScope checks whether a string is a valid Scope.
func IsValidScope(s string) bool {
	for _, valid := range ValidScopes {
		if Scope(s) == valid {
			return true
		}
	}
	return false
}

// ValidSearchFields contains all fields that can be targeted in FTS5 search.
var ValidSearchFields = []string{"title", "what", "why", "location", "learned", "tags"}

// IsValidSearchField checks whether a string is a valid search field.
func IsValidSearchField(f string) bool {
	for _, valid := range ValidSearchFields {
		if f == valid {
			return true
		}
	}
	return false
}

// Memory represents a unit of persistent knowledge.
type Memory struct {
	ID        int64
	Title     string
	Type      MemoryType
	Project   string
	Scope     Scope
	What      string
	Why       string
	Location  string
	Learned   string
	Tags      []string
	TopicKey  string
	SessionID string
	Source    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SearchQuery defines parameters for searching memories.
type SearchQuery struct {
	Text      string
	Type      MemoryType
	Project   string
	Scope     Scope
	Field     string
	SessionID string
	Limit     int
}

// SearchResult wraps a memory with its relevance score.
type SearchResult struct {
	Memory Memory
	Rank   float64
}

// UpdateParams holds optional fields for updating a memory.
// Nil pointers mean "do not update this field".
type UpdateParams struct {
	Title    *string
	Type     *MemoryType
	What     *string
	Why      *string
	Location *string
	Learned  *string
	Tags     *[]string
	TopicKey *string
}

// Stats holds aggregate statistics about stored memories.
type Stats struct {
	TotalMemories int
	ByType        map[MemoryType]int
	ByProject     map[string]int
	OldestMemory  time.Time
	NewestMemory  time.Time
}
