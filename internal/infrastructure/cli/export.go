package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var project, file string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export memories to JSON",
		Long:  "Export all memories (or filtered by project) to a JSON file or stdout.",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			data, err := d.export.Execute(cmd.Context(), project)
			if err != nil {
				return err
			}

			if file != "" {
				if err := os.WriteFile(file, data, 0644); err != nil {
					return fmt.Errorf("writing file %s: %w", file, err)
				}
				fmt.Printf("Exported to %s\n", file)
			} else {
				fmt.Println(string(data))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Filter by project")
	cmd.Flags().StringVar(&file, "file", "", "Output file (default: stdout)")

	return cmd
}
