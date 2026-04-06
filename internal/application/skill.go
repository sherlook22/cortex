package application

import (
	"fmt"
	"strings"
)

// GenerateSkillUseCase generates a markdown skill document describing the CLI.
type GenerateSkillUseCase struct {
	version string
}

// NewGenerateSkillUseCase creates a new GenerateSkillUseCase.
func NewGenerateSkillUseCase(version string) *GenerateSkillUseCase {
	return &GenerateSkillUseCase{version: version}
}

// Execute generates the skill markdown content.
func (uc *GenerateSkillUseCase) Execute() string {
	var sb strings.Builder

	sb.WriteString("# Cortex Memory CLI\n\n")
	sb.WriteString("Persistent memory system for AI coding agents. All data is stored in SQLite with full-text search.\n\n")
	sb.WriteString(fmt.Sprintf("Version: %s\n\n", uc.version))

	sb.WriteString("## Commands\n\n")

	sb.WriteString("### cortex save\n")
	sb.WriteString("Save a structured memory.\n")
	sb.WriteString("- `--title` (required): Short, searchable title\n")
	sb.WriteString("- `--type` (required): bugfix | decision | architecture | discovery | pattern | config\n")
	sb.WriteString("- `--project` (required): Project name\n")
	sb.WriteString("- `--what` (required): What was done\n")
	sb.WriteString("- `--why` (required): Why it was done\n")
	sb.WriteString("- `--where` (required): Affected files/paths\n")
	sb.WriteString("- `--learned` (required): What was learned\n")
	sb.WriteString("- `--scope`: project (default) | personal\n")
	sb.WriteString("- `--tags`: Comma-separated tags\n")
	sb.WriteString("- `--topic-key`: Stable key for upserts (e.g. architecture/auth). Same key updates existing memory instead of creating new.\n")
	sb.WriteString("- `--session`: Session ID (auto-generated if omitted)\n\n")

	sb.WriteString("### cortex search <query>\n")
	sb.WriteString("Full-text search across all memory fields.\n")
	sb.WriteString("- `<query>` (positional, required): Search terms\n")
	sb.WriteString("- `--type`: Filter by memory type\n")
	sb.WriteString("- `--project`: Filter by project\n")
	sb.WriteString("- `--scope`: Filter by scope\n")
	sb.WriteString("- `--field`: Search in specific field (title, what, why, location, learned, tags)\n")
	sb.WriteString("- `--session`: Filter by session ID\n")
	sb.WriteString("- `--limit`: Max results (default 10)\n\n")

	sb.WriteString("### cortex get <id>\n")
	sb.WriteString("Retrieve full details of a memory by its numeric ID.\n\n")

	sb.WriteString("### cortex update <id>\n")
	sb.WriteString("Update specific fields of an existing memory. Only provided flags are modified.\n")
	sb.WriteString("- `--title`: New title\n")
	sb.WriteString("- `--type`: New type\n")
	sb.WriteString("- `--what`: New what\n")
	sb.WriteString("- `--why`: New why\n")
	sb.WriteString("- `--where`: New location\n")
	sb.WriteString("- `--learned`: New learned\n")
	sb.WriteString("- `--tags`: New tags (comma-separated)\n")
	sb.WriteString("- `--topic-key`: New topic key\n\n")

	sb.WriteString("### cortex delete <id>\n")
	sb.WriteString("Permanently delete a memory.\n")
	sb.WriteString("- `--force`: Skip confirmation prompt\n\n")

	sb.WriteString("### cortex context\n")
	sb.WriteString("Get recent memories formatted as readable context.\n")
	sb.WriteString("- `--project`: Filter by project\n")
	sb.WriteString("- `--session`: Filter by session ID\n")
	sb.WriteString("- `--limit`: Max memories (default 20)\n\n")

	sb.WriteString("### cortex stats\n")
	sb.WriteString("Show aggregate statistics.\n")
	sb.WriteString("- `--project`: Filter by project\n\n")

	sb.WriteString("### cortex export\n")
	sb.WriteString("Export all memories to JSON.\n")
	sb.WriteString("- `--project`: Filter by project\n")
	sb.WriteString("- `--file`: Output file (default: stdout)\n\n")

	sb.WriteString("### cortex import\n")
	sb.WriteString("Import memories from a JSON file.\n")
	sb.WriteString("- `--file` (required): Input file\n\n")

	sb.WriteString("### cortex session\n")
	sb.WriteString("Manage agent sessions.\n\n")

	sb.WriteString("#### cortex session start\n")
	sb.WriteString("Create or reopen a session. Idempotent.\n")
	sb.WriteString("- `--id` (required): Session ID\n")
	sb.WriteString("- `--project` (required): Project name\n")
	sb.WriteString("- `--directory`: Working directory\n\n")

	sb.WriteString("#### cortex session end\n")
	sb.WriteString("Close a session with a summary.\n")
	sb.WriteString("- `--id` (required): Session ID\n")
	sb.WriteString("- `--summary`: Session summary\n\n")

	sb.WriteString("#### cortex session list\n")
	sb.WriteString("List recent sessions.\n")
	sb.WriteString("- `--project`: Filter by project\n")
	sb.WriteString("- `--limit`: Max results (default 10)\n\n")

	sb.WriteString("#### cortex session get <id>\n")
	sb.WriteString("Get session details.\n\n")

	sb.WriteString("### cortex capture\n")
	sb.WriteString("Capture learnings from raw text via stdin.\n")
	sb.WriteString("- `--project` (required): Project name\n")
	sb.WriteString("- `--session`: Session ID (auto-generated if omitted)\n")
	sb.WriteString("- `--source`: Origin (default: manual). Use 'subagent' for subagent output.\n")
	sb.WriteString("\nExample:\n")
	sb.WriteString("```\necho \"subagent output...\" | cortex capture --project myapp --source subagent\n```\n\n")

	sb.WriteString("## Global Flags\n\n")
	sb.WriteString("- `--output-json`: Output in JSON format (useful for scripting)\n\n")

	sb.WriteString("## When to Save (recommended triggers)\n\n")
	sb.WriteString("- After completing a bug fix\n")
	sb.WriteString("- After making an architecture or design decision\n")
	sb.WriteString("- After discovering something non-obvious about the codebase\n")
	sb.WriteString("- After a configuration change\n")
	sb.WriteString("- After establishing a new pattern or convention\n\n")

	sb.WriteString("## When to Search\n\n")
	sb.WriteString("- Before starting work that might overlap with past work\n")
	sb.WriteString("- When the user asks to recall something from a previous session\n")
	sb.WriteString("- When working on a topic you have no context on\n\n")

	sb.WriteString("## Session Protocol\n\n")
	sb.WriteString("1. **On start**: Run `cortex session start --id SESSION --project PROJECT --directory DIR`\n")
	sb.WriteString("2. **During work**: Use `--session SESSION` with `cortex save` to associate memories\n")
	sb.WriteString("3. **On end**: Save a session summary, then `cortex session end --id SESSION --summary \"...\"`\n")
	sb.WriteString("4. **After compaction**: Save compacted summary, run `cortex context --project PROJECT --session SESSION`, continue\n\n")

	sb.WriteString("## Topic Key Convention\n\n")
	sb.WriteString("Use `{category}/{topic}` format for evolving knowledge:\n")
	sb.WriteString("- `architecture/auth-model` — authentication design decisions\n")
	sb.WriteString("- `bug/n-plus-1-users` — specific bug investigation\n")
	sb.WriteString("- `config/docker-setup` — environment configuration\n")
	sb.WriteString("- `pattern/error-handling` — established conventions\n")

	return sb.String()
}
