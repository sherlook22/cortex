package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sherlook22/cortex/internal/application"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var req application.SearchMemoryRequest

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search memories by text",
		Long:  "Full-text search across all memory fields with optional filters.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			req.Text = strings.Join(args, " ")

			results, err := d.search.Execute(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.MarshalIndent(results, "", "  ")
				fmt.Println(string(out))
			} else {
				if len(results) == 0 {
					fmt.Println("No memories found.")
					return nil
				}
				for _, r := range results {
					m := r.Memory
					fmt.Printf("[%d] %s (%s, %s, %s)\n",
						m.ID, m.Title, m.Type, m.Project,
						m.CreatedAt.Format("2006-01-02"))
					fmt.Printf("     What: %s\n", truncate(m.What, 80))
					fmt.Println()
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&req.Type, "type", "", "Filter by type")
	cmd.Flags().StringVar(&req.Project, "project", "", "Filter by project")
	cmd.Flags().StringVar(&req.Scope, "scope", "", "Filter by scope")
	cmd.Flags().StringVar(&req.Field, "field", "", "Search in specific field (title, what, why, location, learned, tags)")
	cmd.Flags().IntVar(&req.Limit, "limit", 10, "Maximum results")

	return cmd
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
