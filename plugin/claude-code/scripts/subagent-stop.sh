#!/bin/bash
set -euo pipefail

# Subagent stop hook for Claude Code.
# Reads hook input from stdin, captures subagent output for passive learning extraction.

# Read hook input from stdin (JSON with session_id, cwd, stdout fields).
INPUT=$(cat)

# Extract fields from JSON input.
SESSION_ID=$(echo "$INPUT" | grep -oP '"session_id"\s*:\s*"\K[^"]*' 2>/dev/null || true)
CWD=$(echo "$INPUT" | grep -oP '"cwd"\s*:\s*"\K[^"]*' 2>/dev/null || true)
CWD="${CWD:-$(pwd)}"
PROJECT=$(basename "$CWD")

# Extract stdout from JSON input (use jq if available, fallback to grep).
if command -v jq > /dev/null 2>&1; then
  STDOUT=$(echo "$INPUT" | jq -r '.stdout // empty' 2>/dev/null || true)
else
  STDOUT=$(echo "$INPUT" | grep -oP '"stdout"\s*:\s*"\K[^"]*' 2>/dev/null || true)
fi

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
