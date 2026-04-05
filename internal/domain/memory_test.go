package domain

import "testing"

func TestIsValidMemoryType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "valid bugfix", input: "bugfix", expected: true},
		{name: "valid decision", input: "decision", expected: true},
		{name: "valid architecture", input: "architecture", expected: true},
		{name: "valid discovery", input: "discovery", expected: true},
		{name: "valid pattern", input: "pattern", expected: true},
		{name: "valid config", input: "config", expected: true},
		{name: "invalid empty", input: "", expected: false},
		{name: "invalid random", input: "foobar", expected: false},
		{name: "invalid uppercase", input: "Bugfix", expected: false},
		{name: "invalid with spaces", input: "bug fix", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidMemoryType(tt.input)
			if got != tt.expected {
				t.Errorf("IsValidMemoryType(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidScope(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "valid project", input: "project", expected: true},
		{name: "valid personal", input: "personal", expected: true},
		{name: "invalid empty", input: "", expected: false},
		{name: "invalid random", input: "global", expected: false},
		{name: "invalid uppercase", input: "Project", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidScope(tt.input)
			if got != tt.expected {
				t.Errorf("IsValidScope(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
