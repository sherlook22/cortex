#!/bin/bash
set -euo pipefail

# Subagent stop hook for Claude Code.
# Captures subagent output for passive learning extraction.

CWD="${CWD:-$(pwd)}"
PROJECT=$(basename "$CWD")
SESSION_ID="${SESSION_ID:-}"

# Read hook input from stdin (JSON with session_id, cwd, stdout fields).
INPUT=$(cat)

# Extract stdout from JSON input.
STDOUT=$(echo "$INPUT" | jq -r '.stdout // empty' 2>/dev/null || true)

# Exit early if no output.
if [ -z "$STDOUT" ] || [ ${#STDOUT} -lt 50 ]; then
  exit 0
fi

# Pipe to cortex capture.
if [ -n "$SESSION_ID" ]; then
  echo "$STDOUT" | cortex capture --project "$PROJECT" --session "$SESSION_ID" --source subagent 2>/dev/null || true
else
  echo "$STDOUT" | cortex capture --project "$PROJECT" --source subagent 2>/dev/null || true
fi
