package cli

import (
	"fmt"

	"github.com/sherlook22/cortex/internal/application"
	"github.com/spf13/cobra"
)

func newSkillCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Generate skill markdown for AI agents",
		Long:  "Generate a markdown document describing all CLI commands, designed to be loaded as a skill by AI coding agents.",
		Run: func(cmd *cobra.Command, args []string) {
			uc := application.NewGenerateSkillUseCase(version)
			fmt.Print(uc.Execute())
		},
	}

	return cmd
}
