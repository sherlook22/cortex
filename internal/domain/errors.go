package domain

import "errors"

var (
	// ErrMemoryNotFound is returned when a memory with the given ID does not exist.
	ErrMemoryNotFound = errors.New("memory not found")

	// ErrInvalidMemoryType is returned when an unrecognized memory type is provided.
	ErrInvalidMemoryType = errors.New("invalid memory type")

	// ErrInvalidScope is returned when an unrecognized scope is provided.
	ErrInvalidScope = errors.New("invalid scope")

	// ErrEmptyTitle is returned when a memory title is empty.
	ErrEmptyTitle = errors.New("title is required")

	// ErrEmptyProject is returned when a memory project is empty.
	ErrEmptyProject = errors.New("project is required")

	// ErrEmptyWhat is returned when the what field is empty.
	ErrEmptyWhat = errors.New("what is required")

	// ErrEmptyWhy is returned when the why field is empty.
	ErrEmptyWhy = errors.New("why is required")

	// ErrEmptyLocation is returned when the location field is empty.
	ErrEmptyLocation = errors.New("location is required")

	// ErrEmptyLearned is returned when the learned field is empty.
	ErrEmptyLearned = errors.New("learned is required")

	// ErrEmptySearchQuery is returned when search text is empty.
	ErrEmptySearchQuery = errors.New("search query is required")
)
