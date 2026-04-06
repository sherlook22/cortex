package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSkillUseCase_Execute(t *testing.T) {
	expectedVersion := "1.0.0"
	expectedCommands := []string{
		"cortex save", "cortex search", "cortex get",
		"cortex update", "cortex delete", "cortex context",
		"cortex stats", "cortex export", "cortex import",
	}

	testCases := []struct {
		name   string
		args   func() string
		assert func(t *testing.T, result string)
	}{
		{
			name: "includes version",
			args: func() string { return expectedVersion },
			assert: func(t *testing.T, result string) {
				assert.Contains(t, result, expectedVersion)
			},
		},
		{
			name: "includes all commands",
			args: func() string { return expectedVersion },
			assert: func(t *testing.T, result string) {
				for _, cmd := range expectedCommands {
					assert.Contains(t, result, cmd)
				}
			},
		},
		{
			name: "includes required flags for save",
			args: func() string { return expectedVersion },
			assert: func(t *testing.T, result string) {
				requiredFlags := []string{"--title", "--type", "--project", "--what", "--why", "--where", "--learned"}
				for _, flag := range requiredFlags {
					assert.Contains(t, result, flag)
				}
			},
		},
		{
			name: "includes when to save guidelines",
			args: func() string { return expectedVersion },
			assert: func(t *testing.T, result string) {
				assert.Contains(t, result, "When to Save")
			},
		},
		{
			name: "includes topic key convention",
			args: func() string { return expectedVersion },
			assert: func(t *testing.T, result string) {
				assert.Contains(t, result, "Topic Key Convention")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			version := tc.args()

			uc := NewGenerateSkillUseCase(version)
			result := uc.Execute()

			tc.assert(t, result)
		})
	}
}
