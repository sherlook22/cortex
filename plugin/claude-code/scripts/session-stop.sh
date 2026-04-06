#!/bin/bash
set -euo pipefail

# Session stop hook for Claude Code.
# Safety net: ensures the session record is closed even if the agent didn't
# explicitly run `cortex session end` with a summary before the session ended.
# The agent is expected to have already ended the session with a summary via
# the Session Close Protocol; this hook is only a fallback to avoid orphaned sessions.

# --- Dependencies ---
if ! command -v jq > /dev/null 2>&1; then
  exit 0
fi

# --- Parse JSON input from stdin ---
INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')

if [ -n "$SESSION_ID" ]; then
  cortex session end --id "$SESSION_ID" 2>/dev/null || true
fi
