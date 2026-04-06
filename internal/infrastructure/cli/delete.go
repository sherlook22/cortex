package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a memory",
		Long:  "Permanently delete a memory by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid ID: %s", args[0])
			}

			if !force {
				fmt.Printf("Are you sure you want to delete memory #%d? [y/N] ", id)
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			if err := d.del.Execute(cmd.Context(), id); err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.Marshal(map[string]any{"id": id, "status": "deleted"})
				fmt.Println(string(out))
			} else {
				fmt.Printf("Deleted memory #%d\n", id)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
