#!/bin/bash
set -euo pipefail

# Post-compaction hook for Claude Code.
# Injects previous context and recovery instruction.

SESSION_ID="${SESSION_ID:-}"
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

cat <<PROTOCOL
## Cortex Memory — Post-Compaction Recovery

CRITICAL INSTRUCTION POST-COMPACTION:
FIRST ACTION REQUIRED: Save the compacted summary above as a memory using:
  cortex save --title "Session summary (compacted)" --type decision --project ${PROJECT} --session ${SESSION_ID} --what "..." --why "Compaction recovery" --where "..." --learned "..."

Then continue working.

${CONTEXT}
PROTOCOL
