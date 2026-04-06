package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a memory by ID",
		Long:  "Retrieve the full details of a specific memory.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid ID: %s", args[0])
			}

			memory, err := d.get.Execute(cmd.Context(), id)
			if err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.MarshalIndent(memory, "", "  ")
				fmt.Println(string(out))
			} else {
				fmt.Printf("# [%d] %s\n\n", memory.ID, memory.Title)
				fmt.Printf("- **Type**: %s\n", memory.Type)
				fmt.Printf("- **Project**: %s\n", memory.Project)
				fmt.Printf("- **Scope**: %s\n", memory.Scope)
				fmt.Printf("- **Created**: %s\n", memory.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("- **Updated**: %s\n", memory.UpdatedAt.Format("2006-01-02 15:04:05"))
				if memory.TopicKey != "" {
					fmt.Printf("- **Topic Key**: %s\n", memory.TopicKey)
				}
				if len(memory.Tags) > 0 {
					fmt.Printf("- **Tags**: %s\n", strings.Join(memory.Tags, ", "))
				}
				fmt.Println()
				fmt.Printf("**What**: %s\n", memory.What)
				fmt.Printf("**Why**: %s\n", memory.Why)
				fmt.Printf("**Location**: %s\n", memory.Location)
				fmt.Printf("**Learned**: %s\n", memory.Learned)
			}
			return nil
		},
	}

	return cmd
}
