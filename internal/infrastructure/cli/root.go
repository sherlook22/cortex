package cli

import (
	"github.com/spf13/cobra"
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cortex",
		Short:   "Persistent memory system for AI coding agents",
		Long:    "Cortex is a CLI tool that provides persistent memory storage with full-text search, designed for AI coding agents and developers.",
		Version: version,
	}

	return cmd
}

// Execute runs the root command.
func Execute(version string) error {
	return newRootCmd(version).Execute()
}
