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
		Long:  "Registers cortex as a local marketplace and installs the plugin via claude CLI.",
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
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	// Extract plugin to a permanent marketplace directory.
	// Structure: ~/.cortex/marketplace/.claude-plugin/marketplace.json
	//            ~/.cortex/marketplace/plugins/cortex/{plugin files}
	marketplaceDir := filepath.Join(home, ".cortex", "marketplace")
	pluginDir := filepath.Join(marketplaceDir, "plugins", "cortex")

	// Remove old installation if present to ensure clean state.
	os.RemoveAll(pluginDir)

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

	// Write marketplace manifest.
	manifestDir := filepath.Join(marketplaceDir, ".claude-plugin")
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("creating manifest dir: %w", err)
	}

	manifest := map[string]any{
		"name":        "cortex-marketplace",
		"description": "Cortex persistent memory plugins",
		"owner": map[string]string{
			"name":  "cortex",
			"email": "cortex@local",
		},
		"plugins": []map[string]any{
			{
				"name":        "cortex",
				"description": "Persistent memory system for AI coding agents",
				"source":      "./plugins/cortex",
				"category":    "productivity",
			},
		},
	}
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(manifestDir, "marketplace.json"), manifestData, 0644); err != nil {
		return fmt.Errorf("writing marketplace manifest: %w", err)
	}

	// Find claude CLI.
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI not found in PATH: %w", err)
	}

	// Register as local marketplace (idempotent — ignores if already added).
	addMkt := exec.Command(claudeBin, "plugin", "marketplace", "add", marketplaceDir)
	addMkt.Stdout = os.Stdout
	addMkt.Stderr = os.Stderr
	if err := addMkt.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: marketplace may already be registered: %v\n", err)
	}

	// Install the plugin from the marketplace.
	installCmd := exec.Command(claudeBin, "plugin", "install", "cortex")
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
