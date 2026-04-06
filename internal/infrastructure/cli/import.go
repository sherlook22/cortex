package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import memories from JSON",
		Long:  "Import memories from a JSON file previously created with 'cortex export'.",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("reading file %s: %w", file, err)
			}

			count, err := d.imp.Execute(cmd.Context(), data)
			if err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.Marshal(map[string]any{"imported": count})
				fmt.Println(string(out))
			} else {
				fmt.Printf("Imported %d memories\n", count)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Input JSON file (required)")
	cmd.MarkFlagRequired("file")

	return cmd
}
