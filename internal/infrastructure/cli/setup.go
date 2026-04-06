package cli

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	pluginfs "github.com/sherlook22/cortex/plugin"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup <agent>",
		Short: "Configure cortex for an AI agent",
		Long:  "Install the cortex plugin for a supported AI agent.\n\nSupported agents: claude-code, opencode",
	}

	cmd.AddCommand(
		newSetupClaudeCodeCmd(),
		newSetupOpenCodeCmd(),
	)

	return cmd
}

func newSetupClaudeCodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claude-code",
		Short: "Install cortex plugin for Claude Code",
		Long:  "Extracts the plugin to a temporary directory and installs it via claude plugin install.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setupClaudeCode()
		},
	}
}

func newSetupOpenCodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "opencode",
		Short: "Install cortex plugin for OpenCode",
		Long:  "Copies the cortex plugin to the OpenCode plugins directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setupOpenCode()
		},
	}
}

func setupClaudeCode() error {
	// Extract embedded plugin to a temp directory.
	tmpDir, err := os.MkdirTemp("", "cortex-claude-plugin-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pluginDir := filepath.Join(tmpDir, "claude-code")
	if err := extractEmbeddedFS(pluginfs.ClaudeCodeFS, "claude-code", pluginDir); err != nil {
		return fmt.Errorf("extracting plugin: %w", err)
	}

	// Make scripts executable.
	scriptsDir := filepath.Join(pluginDir, "scripts")
	entries, _ := os.ReadDir(scriptsDir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sh") {
			os.Chmod(filepath.Join(scriptsDir, e.Name()), 0755)
		}
	}

	// Install via claude CLI.
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI not found in PATH: %w", err)
	}

	installCmd := exec.Command(claudeBin, "plugin", "install", "--plugin-dir", pluginDir)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("installing plugin: %w", err)
	}

	// Add cortex to Bash allowlist.
	if err := addClaudeCodeAllowlist(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not update allowlist: %v\n", err)
	}

	fmt.Println("Cortex plugin installed for Claude Code.")
	return nil
}

func setupOpenCode() error {
	configDir := openCodeConfigDir()
	pluginsDir := filepath.Join(configDir, "plugins")

	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return fmt.Errorf("creating plugins dir: %w", err)
	}

	pluginPath := filepath.Join(pluginsDir, "cortex.ts")
	if err := os.WriteFile(pluginPath, pluginfs.OpenCodePlugin, 0644); err != nil {
		return fmt.Errorf("writing plugin file: %w", err)
	}

	fmt.Printf("Cortex plugin installed at %s\n", pluginPath)
	return nil
}

// extractEmbeddedFS extracts files from an embed.FS to disk.
func extractEmbeddedFS(fsys fs.FS, root string, dest string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(root, path)
		targetPath := filepath.Join(dest, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("reading embedded file %s: %w", path, err)
		}

		return os.WriteFile(targetPath, data, 0644)
	})
}

// addClaudeCodeAllowlist adds cortex to the Claude Code Bash command allowlist.
func addClaudeCodeAllowlist() error {
	settingsPath := filepath.Join(claudeConfigDir(), "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create a new settings file with just the allowlist.
			settings := map[string]any{
				"permissions": map[string]any{
					"allow": []string{"Bash(cortex *)"},
				},
			}
			out, _ := json.MarshalIndent(settings, "", "  ")
			return os.WriteFile(settingsPath, out, 0644)
		}
		return err
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parsing settings.json: %w", err)
	}

	perms, ok := settings["permissions"].(map[string]any)
	if !ok {
		perms = map[string]any{}
		settings["permissions"] = perms
	}

	allowList, ok := perms["allow"].([]any)
	if !ok {
		allowList = []any{}
	}

	// Check if already in allowlist.
	cortexPattern := "Bash(cortex *)"
	for _, item := range allowList {
		if s, ok := item.(string); ok && s == cortexPattern {
			return nil // Already present.
		}
	}

	allowList = append(allowList, cortexPattern)
	perms["allow"] = allowList

	out, _ := json.MarshalIndent(settings, "", "  ")
	return os.WriteFile(settingsPath, out, 0644)
}

func claudeConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

func openCodeConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "opencode")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode")
}
