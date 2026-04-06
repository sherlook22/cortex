#!/bin/bash
set -euo pipefail

# Post-compaction hook for Claude Code.
# Reads hook input from stdin, injects previous context and recovery instruction.

# --- Dependencies ---
if ! command -v jq > /dev/null 2>&1; then
  echo "cortex plugin: jq is required but not found." >&2
  exit 0
fi

if ! command -v python3 > /dev/null 2>&1; then
  echo "cortex plugin: python3 is required but not found." >&2
  exit 0
fi

# --- Parse JSON input from stdin ---
INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')
CWD=$(echo "$INPUT" | jq -r '.cwd // empty')
CWD="${CWD:-$(pwd)}"
PROJECT=$(basename "$CWD")
COMPACT_SUMMARY=$(echo "$INPUT" | jq -r '.compact_summary // empty')

# Ensure session exists (may have been lost during compaction). Suppress all output.
if [ -n "$SESSION_ID" ]; then
  cortex session start --id "$SESSION_ID" --project "$PROJECT" --directory "$CWD" >/dev/null 2>&1 || true
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

# Build recovery message, embedding the compact summary so the agent has it available.
RECOVERY="## Cortex Memory — Post-Compaction Recovery

CRITICAL INSTRUCTION POST-COMPACTION:
FIRST ACTION REQUIRED: Save the compacted summary as a memory using:
  cortex save --title \"Session summary (compacted)\" --type decision --project ${PROJECT}${SESSION_FLAG} --what \"...\" --why \"Compaction recovery\" --where \"...\" --learned \"...\""

if [ -n "$COMPACT_SUMMARY" ]; then
  RECOVERY="${RECOVERY}

### Compacted Summary

${COMPACT_SUMMARY}"
fi

RECOVERY="${RECOVERY}

Then continue working."

if [ -n "$CONTEXT" ]; then
  RECOVERY="${RECOVERY}

### Previous Memories

${CONTEXT}"
fi

# Output as hookSpecificOutput JSON.
ESCAPED=$(echo "$RECOVERY" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read()))")
cat <<HOOK_JSON
{
  "hookSpecificOutput": {
    "hookEventName": "PostCompact",
    "additionalContext": ${ESCAPED}
  }
}
HOOK_JSON
