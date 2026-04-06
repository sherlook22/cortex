#!/bin/bash
set -euo pipefail

# Post-compaction hook for Claude Code.
# Reads hook input from stdin, injects previous context and recovery instruction.

# Parse JSON input from stdin.
INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | grep -oP '"session_id"\s*:\s*"\K[^"]*' 2>/dev/null || true)
CWD=$(echo "$INPUT" | grep -oP '"cwd"\s*:\s*"\K[^"]*' 2>/dev/null || true)
CWD="${CWD:-$(pwd)}"
PROJECT=$(basename "$CWD")

# Ensure session exists (may have been lost during compaction).
if [ -n "$SESSION_ID" ]; then
  cortex session start --id "$SESSION_ID" --project "$PROJECT" --directory "$CWD" 2>/dev/null || true
fi

# Fetch context (session-specific if available, otherwise project-wide).
if [ -n "$SESSION_ID" ]; then
  CONTEXT=$(cortex context --project "$PROJECT" --session "$SESSION_ID" 2>/dev/null || true)
else
  CONTEXT=$(cortex context --project "$PROJECT" 2>/dev/null || true)
fi

SESSION_FLAG=""
if [ -n "$SESSION_ID" ]; then
  SESSION_FLAG=" --session ${SESSION_ID}"
fi

# Build recovery message.
RECOVERY="## Cortex Memory — Post-Compaction Recovery

CRITICAL INSTRUCTION POST-COMPACTION:
FIRST ACTION REQUIRED: Save the compacted summary above as a memory using:
  cortex save --title \"Session summary (compacted)\" --type decision --project ${PROJECT}${SESSION_FLAG} --what \"...\" --why \"Compaction recovery\" --where \"...\" --learned \"...\"

Then continue working.

${CONTEXT}"

# Output as hookSpecificOutput JSON.
ESCAPED=$(echo "$RECOVERY" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read()))" 2>/dev/null || echo '""')
cat <<HOOK_JSON
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": ${ESCAPED}
  }
}
HOOK_JSON
