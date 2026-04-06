#!/bin/bash
set -euo pipefail

# Session start hook for Claude Code.
# Reads hook input from stdin, creates a session, injects Memory Protocol + context.

# --- Dependencies ---
# Require jq for reliable JSON parsing.
if ! command -v jq > /dev/null 2>&1; then
  echo "cortex plugin: jq is required but not found. Install it: https://jqlang.github.io/jq/" >&2
  exit 0
fi

# Require python3 for JSON escaping output.
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

# Start session (idempotent). Suppress all output.
if [ -n "$SESSION_ID" ]; then
  cortex session start --id "$SESSION_ID" --project "$PROJECT" --directory "$CWD" >/dev/null 2>&1 || true
fi

# Build memory context (include session-scoped context on resume).
if [ -n "$SESSION_ID" ]; then
  CONTEXT=$(cortex context --project "$PROJECT" --session "$SESSION_ID" 2>/dev/null || true)
else
  CONTEXT=$(cortex context --project "$PROJECT" 2>/dev/null || true)
fi

# Build the full injection: Memory Protocol + previous context.
SESSION_FLAG=""
if [ -n "$SESSION_ID" ]; then
  SESSION_FLAG=" --session ${SESSION_ID}"
fi

INJECTION="## Cortex Persistent Memory — Protocol

You have access to cortex, a persistent memory CLI tool. Use it proactively.
Always inform the user when you save or search memories (e.g. \"Saving this decision to memory...\" or \"Searching memory for...\").

### WHEN TO SAVE (mandatory — not optional)

Run \`cortex save\` IMMEDIATELY after any of these:
- Bug fix completed
- Architecture or design decision made
- Non-obvious discovery about the codebase
- Configuration change or environment setup
- Pattern established (naming, structure, convention)
- User preference or constraint learned

Format:
\`\`\`
cortex save \\
  --title \"Verb + what\" \\
  --type <bugfix|decision|architecture|discovery|pattern|config> \\
  --project ${PROJECT}${SESSION_FLAG} \\
  --what \"One sentence\" \\
  --why \"What motivated it\" \\
  --where \"Files or paths affected\" \\
  --learned \"Gotchas, edge cases\"
\`\`\`

### WHEN TO SEARCH

When the user asks to recall something, or proactively when starting related work:
\`\`\`
cortex search \"query\" --project ${PROJECT}
\`\`\`

### SESSION CLOSE PROTOCOL (mandatory)

Before ending, save a session summary with cortex save, then run:
\`\`\`
cortex session end --id ${SESSION_ID} --summary \"brief summary\"
\`\`\`

### AFTER COMPACTION

1. Save the compacted summary as a memory
2. Run \`cortex context --project ${PROJECT}${SESSION_FLAG}\` to recover
3. Continue working"

# Append previous memories if available.
if [ -n "$CONTEXT" ]; then
  INJECTION="${INJECTION}

## Previous Memories

${CONTEXT}"
fi

# Output as hookSpecificOutput JSON for Claude Code.
ESCAPED=$(echo "$INJECTION" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read()))")
cat <<HOOK_JSON
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": ${ESCAPED}
  }
}
HOOK_JSON
