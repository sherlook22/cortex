#!/bin/bash
set -euo pipefail

# Session stop hook for Claude Code.
# The Memory Protocol instructs the agent to save a session summary before
# ending, so this hook simply closes the session record.

SESSION_ID="${SESSION_ID:-}"

if [ -n "$SESSION_ID" ]; then
  cortex session end --id "$SESSION_ID" 2>/dev/null || true
fi
