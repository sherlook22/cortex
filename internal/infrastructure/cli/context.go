package cli

import (
	"encoding/json"
	"fmt"

	"github.com/sherlook22/cortex/internal/application"
	"github.com/spf13/cobra"
)

func newContextCmd() *cobra.Command {
	var req application.GetContextRequest

	cmd := &cobra.Command{
		Use:   "context",
		Short: "Get recent memory context",
		Long:  "Retrieve recent memories formatted as readable context, optionally filtered by project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			result, err := d.context.Execute(cmd.Context(), req)
			if err != nil {
				return err
			}

			if result == "" {
				if outputJSON {
					fmt.Println("[]")
				} else {
					fmt.Println("No memories found.")
				}
				return nil
			}

			if outputJSON {
				out, _ := json.Marshal(map[string]any{"context": result})
				fmt.Println(string(out))
			} else {
				fmt.Print(result)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&req.Project, "project", "", "Filter by project")
	cmd.Flags().IntVar(&req.Limit, "limit", 20, "Maximum memories to include")

	return cmd
}
