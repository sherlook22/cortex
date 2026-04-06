package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sherlook22/cortex/internal/application"
	"github.com/spf13/cobra"
)

func newSaveCmd() *cobra.Command {
	var req application.SaveMemoryRequest
	var tagsStr string

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save a new memory",
		Long:  "Save a structured memory with title, type, project, and content fields.",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			if tagsStr != "" {
				req.Tags = splitCSV(tagsStr)
			}

			id, err := d.save.Execute(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.Marshal(map[string]any{"id": id, "status": "saved"})
				fmt.Println(string(out))
			} else {
				fmt.Printf("Saved memory #%d\n", id)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&req.Title, "title", "", "Short, searchable title (required)")
	cmd.Flags().StringVar(&req.Type, "type", "", "Memory type: bugfix|decision|architecture|discovery|pattern|config (required)")
	cmd.Flags().StringVar(&req.Project, "project", "", "Project name (required)")
	cmd.Flags().StringVar(&req.Scope, "scope", "project", "Scope: project|personal")
	cmd.Flags().StringVar(&req.What, "what", "", "What was done (required)")
	cmd.Flags().StringVar(&req.Why, "why", "", "Why it was done (required)")
	cmd.Flags().StringVar(&req.Location, "where", "", "Affected files/paths (required)")
	cmd.Flags().StringVar(&req.Learned, "learned", "", "What was learned (required)")
	cmd.Flags().StringVar(&tagsStr, "tags", "", "Comma-separated tags")
	cmd.Flags().StringVar(&req.TopicKey, "topic-key", "", "Stable key for upserts (e.g. architecture/auth)")
	cmd.Flags().StringVar(&req.SessionID, "session", "", "Session ID (auto-generated if omitted)")
	cmd.Flags().StringVar(&req.Source, "source", "", "Source origin (e.g. manual, subagent)")

	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("project")
	cmd.MarkFlagRequired("what")
	cmd.MarkFlagRequired("why")
	cmd.MarkFlagRequired("where")
	cmd.MarkFlagRequired("learned")

	return cmd
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
