package cli

import (
	"encoding/json"
	"fmt"

	"github.com/sherlook22/cortex/internal/application"
	"github.com/spf13/cobra"
)

func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage agent sessions",
		Long:  "Create, end, list, and inspect agent sessions.",
	}

	cmd.AddCommand(
		newSessionStartCmd(),
		newSessionEndCmd(),
		newSessionListCmd(),
		newSessionGetCmd(),
	)

	return cmd
}

func newSessionStartCmd() *cobra.Command {
	var req application.StartSessionRequest

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new session",
		Long:  "Create or reopen a session. Idempotent: if the session already exists, returns without error.",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			if err := d.sessionStart.Execute(cmd.Context(), req); err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.Marshal(map[string]any{"id": req.ID, "status": "started"})
				fmt.Println(string(out))
			} else {
				fmt.Printf("Session %s started\n", req.ID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&req.ID, "id", "", "Session ID (required)")
	cmd.Flags().StringVar(&req.Project, "project", "", "Project name (required)")
	cmd.Flags().StringVar(&req.Directory, "directory", "", "Working directory")

	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("project")

	return cmd
}

func newSessionEndCmd() *cobra.Command {
	var req application.EndSessionRequest

	cmd := &cobra.Command{
		Use:   "end",
		Short: "End a session",
		Long:  "Close a session and store its summary.",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			if err := d.sessionEnd.Execute(cmd.Context(), req); err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.Marshal(map[string]any{"id": req.ID, "status": "completed"})
				fmt.Println(string(out))
			} else {
				fmt.Printf("Session %s ended\n", req.ID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&req.ID, "id", "", "Session ID (required)")
	cmd.Flags().StringVar(&req.Summary, "summary", "", "Session summary")

	cmd.MarkFlagRequired("id")

	return cmd
}

func newSessionListCmd() *cobra.Command {
	var req application.ListSessionsRequest

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent sessions",
		Long:  "List recent sessions, optionally filtered by project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			sessions, err := d.sessionList.Execute(cmd.Context(), req)
			if err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.MarshalIndent(sessions, "", "  ")
				fmt.Println(string(out))
			} else {
				if len(sessions) == 0 {
					fmt.Println("No sessions found.")
					return nil
				}
				for _, s := range sessions {
					fmt.Printf("[%s] %s (%s, %s)\n",
						s.ID, s.Project, s.Status,
						s.CreatedAt.Format("2006-01-02 15:04"))
					if s.Summary != "" {
						fmt.Printf("     Summary: %s\n", truncate(s.Summary, 80))
					}
					fmt.Println()
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&req.Project, "project", "", "Filter by project")
	cmd.Flags().IntVar(&req.Limit, "limit", 10, "Maximum results")

	return cmd
}

func newSessionGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get session details",
		Long:  "Retrieve a single session by its ID.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d := initDeps()
			defer d.close()

			session, err := d.sessionGet.Execute(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if outputJSON {
				out, _ := json.MarshalIndent(session, "", "  ")
				fmt.Println(string(out))
			} else {
				fmt.Printf("ID:        %s\n", session.ID)
				fmt.Printf("Project:   %s\n", session.Project)
				fmt.Printf("Directory: %s\n", session.Directory)
				fmt.Printf("Status:    %s\n", session.Status)
				fmt.Printf("Created:   %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("Updated:   %s\n", session.UpdatedAt.Format("2006-01-02 15:04:05"))
				if session.Summary != "" {
					fmt.Printf("Summary:\n%s\n", session.Summary)
				}
			}
			return nil
		},
	}

	return cmd
}
