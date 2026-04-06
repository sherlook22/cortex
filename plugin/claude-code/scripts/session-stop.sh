#!/bin/bash
set -euo pipefail

# Session stop hook for Claude Code.
# Reads hook input from stdin and closes the session record.

# Parse JSON input from stdin.
INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | grep -oP '"session_id"\s*:\s*"\K[^"]*' 2>/dev/null || true)

if [ -n "$SESSION_ID" ]; then
  cortex session end --id "$SESSION_ID" 2>/dev/null || true
fi
