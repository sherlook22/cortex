#!/bin/bash
set -euo pipefail

# Session start hook for Claude Code.
# Creates a session and injects memory context.

SESSION_ID="${SESSION_ID:-}"
CWD="${CWD:-$(pwd)}"
PROJECT=$(basename "$CWD")

# Start session (idempotent).
if [ -n "$SESSION_ID" ]; then
  cortex session start --id "$SESSION_ID" --project "$PROJECT" --directory "$CWD" 2>/dev/null || true
fi

# Inject memory context as additionalContext.
CONTEXT=$(cortex context --project "$PROJECT" 2>/dev/null || true)

if [ -n "$CONTEXT" ]; then
  echo "$CONTEXT"
fi
