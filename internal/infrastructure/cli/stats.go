package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newStatsCmd() *cobra.Command {
	var project string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show memory statistics",
		Long:  "Display aggregate statistics about stored memories.",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			stats, err := d.stats.Execute(cmd.Context(), project)
			if err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.MarshalIndent(stats, "", "  ")
				fmt.Println(string(out))
			} else {
				fmt.Printf("Total memories: %d\n\n", stats.TotalMemories)

				if len(stats.ByType) > 0 {
					fmt.Println("By type:")
					for t, count := range stats.ByType {
						fmt.Printf("  %-15s %d\n", t, count)
					}
					fmt.Println()
				}

				if len(stats.ByProject) > 0 {
					fmt.Println("By project:")
					for p, count := range stats.ByProject {
						fmt.Printf("  %-20s %d\n", p, count)
					}
					fmt.Println()
				}

				if !stats.OldestMemory.IsZero() {
					fmt.Printf("Date range: %s to %s\n",
						stats.OldestMemory.Format("2006-01-02"),
						stats.NewestMemory.Format("2006-01-02"))
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Filter by project")

	return cmd
}
