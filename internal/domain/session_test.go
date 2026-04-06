package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidSessionStatus(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect bool
	}{
		{"valid active", "active", true},
		{"valid completed", "completed", true},
		{"invalid empty", "", false},
		{"invalid random", "paused", false},
		{"invalid uppercase", "Active", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidSessionStatus(tc.input)
			assert.Equal(t, tc.expect, result)
		})
	}
}
