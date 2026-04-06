package domain

import "time"

// SessionStatus represents the lifecycle state of a session.
type SessionStatus string

const (
	SessionActive    SessionStatus = "active"
	SessionCompleted SessionStatus = "completed"
)

// ValidSessionStatuses contains all allowed session statuses.
var ValidSessionStatuses = []SessionStatus{SessionActive, SessionCompleted}

// IsValidSessionStatus checks whether a string is a valid SessionStatus.
func IsValidSessionStatus(s string) bool {
	for _, valid := range ValidSessionStatuses {
		if SessionStatus(s) == valid {
			return true
		}
	}
	return false
}

// Session represents an agent interaction session.
type Session struct {
	ID        string
	Project   string
	Directory string
	Status    SessionStatus
	Summary   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
