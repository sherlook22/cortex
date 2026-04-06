package domain

import "context"

// MemoryRepository defines the port for memory persistence.
// Infrastructure adapters must implement this interface.
type MemoryRepository interface {
	// Save persists a new memory. If TopicKey is set and a memory with the same
	// topic_key/project/scope exists, it performs an upsert.
	// Returns the ID of the saved (or updated) memory.
	Save(ctx context.Context, memory *Memory) (int64, error)

	// Search performs a full-text search across memories.
	Search(ctx context.Context, query SearchQuery) ([]SearchResult, error)

	// GetByID retrieves a single memory by its ID.
	GetByID(ctx context.Context, id int64) (*Memory, error)

	// Update modifies an existing memory. Only non-nil fields in params are updated.
	Update(ctx context.Context, id int64, params UpdateParams) error

	// Delete permanently removes a memory by its ID.
	Delete(ctx context.Context, id int64) error

	// GetRecent returns the most recent memories, optionally filtered by project.
	GetRecent(ctx context.Context, project string, limit int) ([]Memory, error)

	// GetStats returns aggregate statistics, optionally filtered by project.
	GetStats(ctx context.Context, project string) (*Stats, error)

	// FindByTopicKey finds a memory by its topic key within a project and scope.
	FindByTopicKey(ctx context.Context, topicKey string, project string, scope Scope) (*Memory, error)

	// GetAll returns all memories, optionally filtered by project. Used for export.
	GetAll(ctx context.Context, project string) ([]Memory, error)

	// SaveImport persists an imported memory, preserving its original timestamps.
	SaveImport(ctx context.Context, memory *Memory) (int64, error)
}
