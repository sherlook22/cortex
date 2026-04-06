#!/bin/bash
set -euo pipefail

# Session start hook for Claude Code.
# Reads hook input from stdin, creates a session, and injects memory context.

# Parse JSON input from stdin.
INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | grep -oP '"session_id"\s*:\s*"\K[^"]*' 2>/dev/null || true)
CWD=$(echo "$INPUT" | grep -oP '"cwd"\s*:\s*"\K[^"]*' 2>/dev/null || true)
CWD="${CWD:-$(pwd)}"
PROJECT=$(basename "$CWD")

# Start session (idempotent).
if [ -n "$SESSION_ID" ]; then
  cortex session start --id "$SESSION_ID" --project "$PROJECT" --directory "$CWD" 2>/dev/null || true
fi

# Build memory context.
CONTEXT=$(cortex context --project "$PROJECT" 2>/dev/null || true)

# Output as hookSpecificOutput JSON for Claude Code.
if [ -n "$CONTEXT" ]; then
  # Escape the context for JSON embedding.
  ESCAPED=$(echo "$CONTEXT" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read()))" 2>/dev/null || echo '""')
  cat <<HOOK_JSON
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": ${ESCAPED}
  }
}
HOOK_JSON
fi
