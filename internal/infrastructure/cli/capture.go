package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sherlook22/cortex/internal/application"
	"github.com/spf13/cobra"
)

func newCaptureCmd() *cobra.Command {
	var project, sessionID, source string

	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture learnings from raw text",
		Long:  "Read text from stdin, extract structured learnings, and save as memories.\n\nExample:\n  echo \"subagent output...\" | cortex capture --project myapp --source subagent",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}

			req := application.CaptureRequest{
				Input:     string(input),
				Project:   project,
				SessionID: sessionID,
				Source:    source,
			}

			count, err := d.capture.Execute(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.Marshal(map[string]any{"captured": count, "status": "ok"})
				fmt.Println(string(out))
			} else {
				fmt.Printf("Captured %d memory(ies)\n", count)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "Project name (required)")
	cmd.Flags().StringVar(&sessionID, "session", "", "Session ID (auto-generated if omitted)")
	cmd.Flags().StringVar(&source, "source", "manual", "Source: manual, subagent, etc.")

	cmd.MarkFlagRequired("project")

	return cmd
}
