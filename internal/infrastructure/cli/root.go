package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var outputJSON bool

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cortex",
		Short:   "Persistent memory system for AI coding agents",
		Long:    "Cortex is a CLI tool that provides persistent memory storage with full-text search, designed for AI coding agents and developers.",
		Version: version,
	}

	cmd.PersistentFlags().BoolVar(&outputJSON, "output-json", false, "Output in JSON format")

	cmd.AddCommand(
		newSaveCmd(),
		newSearchCmd(),
		newGetCmd(),
		newUpdateCmd(),
		newDeleteCmd(),
		newContextCmd(),
		newStatsCmd(),
		newExportCmd(),
		newImportCmd(),
		newSkillCmd(version),
		newVersionCmd(version),
	)

	return cmd
}

// Execute runs the root command.
func Execute(version string) error {
	return newRootCmd(version).Execute()
}

// initDeps creates dependencies, printing an error and exiting on failure.
func initDeps() *deps {
	d, err := newDeps()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing: %v\n", err)
		os.Exit(1)
	}
	return d
}
