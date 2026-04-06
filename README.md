# Cortex

Persistent memory for AI coding agents. Single binary, SQLite-backed, full-text search.

## Install

```bash
curl -sSL https://raw.githubusercontent.com/sherlook22/cortex/main/scripts/install.sh | sh
```

Or with Go:

```bash
go install github.com/sherlook22/cortex/cmd/cortex@latest
```

## Agent Setup

```bash
# Claude Code
cortex setup claude-code

# OpenCode
cortex setup opencode
```

This installs hooks, scripts, and the memory protocol skill for the agent.

## Commands

```bash
# Save a memory
cortex save --title "Fixed JWT expiration" --type bugfix --project myapp \
  --what "Set token expiry to 24h" --why "Users logged out too frequently" \
  --where "src/auth/token.go:42" --learned "Always configure expiry explicitly" \
  --session my-session-id

# Search
cortex search "auth token"
cortex search "auth" --type bugfix --project myapp --field location

# Read
cortex get 42
cortex context --project myapp --session my-session-id

# Modify
cortex update 42 --title "New title" --learned "New insight"
cortex delete 42 --force

# Sessions
cortex session start --id sess-1 --project myapp --directory /home/dev/myapp
cortex session end --id sess-1 --summary "Implemented auth flow"
cortex session list --project myapp
cortex session get sess-1

# Capture learnings from raw text
echo "subagent output..." | cortex capture --project myapp --source subagent

# Stats & export
cortex stats --project myapp
cortex export --file backup.json
cortex import --file backup.json

# Generate skill doc for AI agents
cortex skill > SKILL.md
```

## Topic Key Upserts

Save with `--topic-key` to update existing knowledge instead of duplicating:

```bash
cortex save --title "Auth strategy" --type decision --project myapp \
  --what "Using JWT" --why "Stateless" --where "src/auth/" \
  --learned "No revocation" --topic-key "architecture/auth"

# Same topic key -> updates in place
cortex save --title "Auth strategy v2" --type decision --project myapp \
  --what "JWT + refresh tokens" --why "Need revocation" --where "src/auth/" \
  --learned "Rotate refresh tokens" --topic-key "architecture/auth"
```

## Memory Types

`bugfix` · `decision` · `architecture` · `discovery` · `pattern` · `config`

## Architecture

Hexagonal architecture. Domain has zero external dependencies.

```
cmd/cortex/                          Entry point
internal/
  domain/                            Entities, ports, value objects
  application/                       Use cases
  infrastructure/storage/sqlite/     SQLite + FTS5 repository
  infrastructure/cli/                Cobra commands
  tests/mocks/                       Generated mocks (mockery)
plugin/
  claude-code/                       Claude Code plugin (hooks, scripts, skill)
  opencode/                          OpenCode plugin (TypeScript)
```

Data stored at `~/.cortex/cortex.db`.

## License

MIT
