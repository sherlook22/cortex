package sqlite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFTS(t *testing.T) {
	testCases := []struct {
		name   string
		args   func() string
		assert func(t *testing.T, result string)
	}{
		{
			name: "simple words",
			args: func() string { return "fix auth bug" },
			assert: func(t *testing.T, result string) {
				assert.Equal(t, `"fix" "auth" "bug"`, result)
			},
		},
		{
			name: "single word",
			args: func() string { return "auth" },
			assert: func(t *testing.T, result string) {
				assert.Equal(t, `"auth"`, result)
			},
		},
		{
			name: "empty string",
			args: func() string { return "" },
			assert: func(t *testing.T, result string) {
				assert.Empty(t, result)
			},
		},
		{
			name: "only spaces",
			args: func() string { return "   " },
			assert: func(t *testing.T, result string) {
				assert.Empty(t, result)
			},
		},
		{
			name: "removes existing quotes",
			args: func() string { return `"auth" "bug"` },
			assert: func(t *testing.T, result string) {
				assert.Equal(t, `"auth" "bug"`, result)
			},
		},
		{
			name: "handles special chars",
			args: func() string { return "fix:auth+bug" },
			assert: func(t *testing.T, result string) {
				assert.Equal(t, `"fix:auth+bug"`, result)
			},
		},
		{
			name: "collapses whitespace",
			args: func() string { return "fix   auth   bug" },
			assert: func(t *testing.T, result string) {
				assert.Equal(t, `"fix" "auth" "bug"`, result)
			},
		},
		{
			name: "strips embedded quotes",
			args: func() string { return `he"llo wo"rld` },
			assert: func(t *testing.T, result string) {
				assert.Equal(t, `"hello" "world"`, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.args()

			result := sanitizeFTS(input)

			tc.assert(t, result)
		})
	}
}
