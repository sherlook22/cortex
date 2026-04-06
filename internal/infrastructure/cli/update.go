package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sherlook22/cortex/internal/application"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var title, typ, what, why, where, learned, tags, topicKey string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a memory",
		Long:  "Update specific fields of an existing memory. Only provided flags are modified.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid ID: %s", args[0])
			}

			req := application.UpdateMemoryRequest{ID: id}

			if cmd.Flags().Changed("title") {
				req.Title = &title
			}
			if cmd.Flags().Changed("type") {
				req.Type = &typ
			}
			if cmd.Flags().Changed("what") {
				req.What = &what
			}
			if cmd.Flags().Changed("why") {
				req.Why = &why
			}
			if cmd.Flags().Changed("where") {
				req.Location = &where
			}
			if cmd.Flags().Changed("learned") {
				req.Learned = &learned
			}
			if cmd.Flags().Changed("tags") {
				t := splitCSV(tags)
				req.Tags = &t
			}
			if cmd.Flags().Changed("topic-key") {
				req.TopicKey = &topicKey
			}

			if err := d.update.Execute(cmd.Context(), req); err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.Marshal(map[string]any{"id": id, "status": "updated"})
				fmt.Println(string(out))
			} else {
				fmt.Printf("Updated memory #%d\n", id)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&typ, "type", "", "New type")
	cmd.Flags().StringVar(&what, "what", "", "New what")
	cmd.Flags().StringVar(&why, "why", "", "New why")
	cmd.Flags().StringVar(&where, "where", "", "New location")
	cmd.Flags().StringVar(&learned, "learned", "", "New learned")
	cmd.Flags().StringVar(&tags, "tags", "", "New tags (comma-separated)")
	cmd.Flags().StringVar(&topicKey, "topic-key", "", "New topic key")

	return cmd
}
