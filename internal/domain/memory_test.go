package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidMemoryType(t *testing.T) {
	testCases := []struct {
		name   string
		args   func() string
		assert func(t *testing.T, result bool)
	}{
		{
			name: "valid bugfix",
			args: func() string { return "bugfix" },
			assert: func(t *testing.T, result bool) {
				assert.True(t, result)
			},
		},
		{
			name: "valid decision",
			args: func() string { return "decision" },
			assert: func(t *testing.T, result bool) {
				assert.True(t, result)
			},
		},
		{
			name: "valid architecture",
			args: func() string { return "architecture" },
			assert: func(t *testing.T, result bool) {
				assert.True(t, result)
			},
		},
		{
			name: "valid discovery",
			args: func() string { return "discovery" },
			assert: func(t *testing.T, result bool) {
				assert.True(t, result)
			},
		},
		{
			name: "valid pattern",
			args: func() string { return "pattern" },
			assert: func(t *testing.T, result bool) {
				assert.True(t, result)
			},
		},
		{
			name: "valid config",
			args: func() string { return "config" },
			assert: func(t *testing.T, result bool) {
				assert.True(t, result)
			},
		},
		{
			name: "invalid empty",
			args: func() string { return "" },
			assert: func(t *testing.T, result bool) {
				assert.False(t, result)
			},
		},
		{
			name: "invalid random",
			args: func() string { return "foobar" },
			assert: func(t *testing.T, result bool) {
				assert.False(t, result)
			},
		},
		{
			name: "invalid uppercase",
			args: func() string { return "Bugfix" },
			assert: func(t *testing.T, result bool) {
				assert.False(t, result)
			},
		},
		{
			name: "invalid with spaces",
			args: func() string { return "bug fix" },
			assert: func(t *testing.T, result bool) {
				assert.False(t, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.args()

			result := IsValidMemoryType(input)

			tc.assert(t, result)
		})
	}
}

func TestIsValidScope(t *testing.T) {
	testCases := []struct {
		name   string
		args   func() string
		assert func(t *testing.T, result bool)
	}{
		{
			name: "valid project",
			args: func() string { return "project" },
			assert: func(t *testing.T, result bool) {
				assert.True(t, result)
			},
		},
		{
			name: "valid personal",
			args: func() string { return "personal" },
			assert: func(t *testing.T, result bool) {
				assert.True(t, result)
			},
		},
		{
			name: "invalid empty",
			args: func() string { return "" },
			assert: func(t *testing.T, result bool) {
				assert.False(t, result)
			},
		},
		{
			name: "invalid random",
			args: func() string { return "global" },
			assert: func(t *testing.T, result bool) {
				assert.False(t, result)
			},
		},
		{
			name: "invalid uppercase",
			args: func() string { return "Project" },
			assert: func(t *testing.T, result bool) {
				assert.False(t, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.args()

			result := IsValidScope(input)

			tc.assert(t, result)
		})
	}
}

func TestIsValidSearchField(t *testing.T) {
	testCases := []struct {
		name   string
		args   func() string
		assert func(t *testing.T, result bool)
	}{
		{
			name:   "valid title",
			args:   func() string { return "title" },
			assert: func(t *testing.T, result bool) { assert.True(t, result) },
		},
		{
			name:   "valid what",
			args:   func() string { return "what" },
			assert: func(t *testing.T, result bool) { assert.True(t, result) },
		},
		{
			name:   "valid why",
			args:   func() string { return "why" },
			assert: func(t *testing.T, result bool) { assert.True(t, result) },
		},
		{
			name:   "valid location",
			args:   func() string { return "location" },
			assert: func(t *testing.T, result bool) { assert.True(t, result) },
		},
		{
			name:   "valid learned",
			args:   func() string { return "learned" },
			assert: func(t *testing.T, result bool) { assert.True(t, result) },
		},
		{
			name:   "valid tags",
			args:   func() string { return "tags" },
			assert: func(t *testing.T, result bool) { assert.True(t, result) },
		},
		{
			name:   "invalid field",
			args:   func() string { return "foobar" },
			assert: func(t *testing.T, result bool) { assert.False(t, result) },
		},
		{
			name:   "invalid empty",
			args:   func() string { return "" },
			assert: func(t *testing.T, result bool) { assert.False(t, result) },
		},
		{
			name:   "invalid content (not a real field)",
			args:   func() string { return "content" },
			assert: func(t *testing.T, result bool) { assert.False(t, result) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.args()

			result := IsValidSearchField(input)

			tc.assert(t, result)
		})
	}
}
