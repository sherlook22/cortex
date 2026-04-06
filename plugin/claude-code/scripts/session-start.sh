#!/bin/bash
set -euo pipefail

# Session start hook for Claude Code.
# Reads hook input from stdin, creates a session, injects Memory Protocol + context.

# Parse JSON input from stdin.
INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | grep -oP '"session_id"\s*:\s*"\K[^"]*' 2>/dev/null || true)
CWD=$(echo "$INPUT" | grep -oP '"cwd"\s*:\s*"\K[^"]*' 2>/dev/null || true)
CWD="${CWD:-$(pwd)}"
PROJECT=$(basename "$CWD")

# Start session (idempotent). Suppress all output.
if [ -n "$SESSION_ID" ]; then
  cortex session start --id "$SESSION_ID" --project "$PROJECT" --directory "$CWD" >/dev/null 2>&1 || true
fi

# Build memory context.
CONTEXT=$(cortex context --project "$PROJECT" 2>/dev/null || true)

# Build the full injection: Memory Protocol + previous context.
SESSION_FLAG=""
if [ -n "$SESSION_ID" ]; then
  SESSION_FLAG=" --session ${SESSION_ID}"
fi

INJECTION="## Cortex Persistent Memory — Protocol

You have access to cortex, a persistent memory CLI tool. Use it proactively.

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
ESCAPED=$(echo "$INJECTION" | python3 -c "import sys,json; print(json.dumps(sys.stdin.read()))" 2>/dev/null || echo '""')
cat <<HOOK_JSON
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": ${ESCAPED}
  }
}
HOOK_JSON
