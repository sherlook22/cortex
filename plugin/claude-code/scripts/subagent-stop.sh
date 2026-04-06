#!/bin/bash
set -euo pipefail

# Subagent stop hook for Claude Code.
# Reads hook input from stdin, captures subagent output for passive learning extraction.

# --- Dependencies ---
if ! command -v jq > /dev/null 2>&1; then
  echo "cortex plugin: jq is required but not found." >&2
  exit 0
fi

# --- Parse JSON input from stdin ---
INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')
CWD=$(echo "$INPUT" | jq -r '.cwd // empty')
CWD="${CWD:-$(pwd)}"
PROJECT=$(basename "$CWD")

# Extract last_assistant_message (the subagent's final response).
LAST_MESSAGE=$(echo "$INPUT" | jq -r '.last_assistant_message // empty')

# Exit early if no output or too short.
if [ -z "$LAST_MESSAGE" ] || [ ${#LAST_MESSAGE} -lt 50 ]; then
  exit 0
fi

# Pipe to cortex capture.
if [ -n "$SESSION_ID" ]; then
  echo "$LAST_MESSAGE" | cortex capture --project "$PROJECT" --session "$SESSION_ID" --source subagent 2>/dev/null || true
else
  echo "$LAST_MESSAGE" | cortex capture --project "$PROJECT" --source subagent 2>/dev/null || true
fi
