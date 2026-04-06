# Cortex

Persistent memory system for AI coding agents. A single Go binary backed by SQLite with full-text search.

## What is Cortex?

Cortex is a CLI tool that gives AI coding agents (and developers) persistent memory across sessions. Instead of losing context when a session ends, agents save structured observations that can be searched and retrieved later.

Key design decisions:
- **CLI-first**: No MCP server, no HTTP API, no TUI. One binary, one interface.
- **Structured fields**: Memories have dedicated `what`, `why`, `where`, and `learned` fields instead of free-text blobs. Each field is independently searchable.
- **SQLite + FTS5**: Single file database with full-text search. Zero runtime dependencies.
- **Agent-friendly, human-readable**: Agents interact via CLI flags, humans read formatted output. Optional `--output-json` for scripting.

## Installation

### Quick install (no Go required)

```bash
curl -sSL https://raw.githubusercontent.com/sherlook22/cortex/main/scripts/install.sh | sh
```

Detects your OS and architecture, downloads the pre-built binary, and installs it to `/usr/local/bin`.

### Install a specific version

```bash
curl -sSL https://raw.githubusercontent.com/sherlook22/cortex/main/scripts/install.sh | sh -s v0.1.0
```

### With Go

```bash
go install github.com/sherlook22/cortex/cmd/cortex@latest
```

### From source

```bash
git clone https://github.com/sherlook22/cortex.git
cd cortex
go build -o cortex ./cmd/cortex/
```

## Usage

### Save a memory

All content fields are required. The agent fills them based on context.

```bash
cortex save \
  --title "Fixed JWT expiration" \
  --type bugfix \
  --project myapp \
  --what "Set token expiry to 24h instead of default 1h" \
  --why "Users were getting logged out too frequently" \
  --where "src/auth/token.go:42, src/config/auth.yaml" \
  --learned "Always configure expiry explicitly, don't rely on library defaults"
```

Optional flags: `--scope`, `--tags`, `--topic-key`.

### Search

```bash
# General search across all fields
cortex search "authentication token"

# Filter by type and project
cortex search "auth" --type bugfix --project myapp

# Search in a specific field
cortex search "src/auth" --field location

# Limit results
cortex search "config" --limit 5
```

### Get full details

```bash
cortex get 42
```

### Update

Only provided flags are modified:

```bash
cortex update 42 --title "Updated title" --learned "New insight"
```

### Delete

```bash
cortex delete 42          # asks for confirmation
cortex delete 42 --force  # skip confirmation
```

### Recent context

```bash
cortex context                     # all recent memories
cortex context --project myapp     # filtered by project
cortex context --limit 10          # limit count
```

### Statistics

```bash
cortex stats
cortex stats --project myapp
```

### Export / Import

```bash
cortex export --file backup.json
cortex export --project myapp --file myapp-memories.json

cortex import --file backup.json
```

### Generate skill for AI agents

```bash
cortex skill > SKILL.md
```

Generates a markdown document describing all commands, designed to be loaded as a skill by AI coding agents.

## Topic Key Upserts

When saving with `--topic-key`, Cortex updates the existing memory with the same key instead of creating a new one. Useful for knowledge that evolves:

```bash
# First save creates the memory
cortex save --title "Auth strategy" --type decision --project myapp \
  --what "Using JWT" --why "Stateless" --where "src/auth/" \
  --learned "Simple but no revocation" --topic-key "architecture/auth"

# Second save with same topic key updates it
cortex save --title "Auth strategy v2" --type decision --project myapp \
  --what "Switched to JWT + refresh tokens" --why "Need revocation" \
  --where "src/auth/" --learned "Rotate refresh tokens" \
  --topic-key "architecture/auth"
```

## Memory Types

| Type | Use for |
|------|---------|
| `bugfix` | Bug fixes, what went wrong and how it was resolved |
| `decision` | Technical decisions with rationale |
| `architecture` | Architectural choices and tradeoffs |
| `discovery` | Non-obvious findings about the codebase |
| `pattern` | Established conventions and patterns |
| `config` | Configuration changes and environment setup |

## Architecture

Cortex follows hexagonal architecture:

```
cmd/cortex/          Entry point
internal/
  domain/            Entities, ports (interfaces), value objects, errors
  application/       Use cases (one per operation)
  infrastructure/
    storage/sqlite/  SQLite + FTS5 repository implementation
    cli/             Cobra CLI commands and dependency wiring
```

- **Domain** has zero external dependencies
- **Application** depends only on domain ports
- **Infrastructure** implements ports and wires everything together

## Data Storage

Database location: `~/.cortex/cortex.db`

Single SQLite file with WAL mode, FTS5 full-text search, and automatic schema migrations via `PRAGMA user_version`.

## License

MIT
