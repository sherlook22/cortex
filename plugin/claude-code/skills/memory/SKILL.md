---
name: cortex-memory
description: >
  ALWAYS ACTIVE — Persistent memory via cortex CLI.
  Save decisions, conventions, bugs, and discoveries proactively.
---

## Cortex Persistent Memory — Protocol

You have access to cortex, a persistent memory CLI tool backed by SQLite with full-text search.

### WHEN TO SAVE (mandatory — not optional)

Run `cortex save` IMMEDIATELY after any of these:
- Bug fix completed
- Architecture or design decision made
- Non-obvious discovery about the codebase
- Configuration change or environment setup
- Pattern established (naming, structure, convention)
- User preference or constraint learned

Format:
```
cortex save \
  --title "Verb + what — short, searchable" \
  --type <bugfix|decision|architecture|discovery|pattern|config> \
  --project PROJECT \
  --session SESSION \
  --what "One sentence — what was done" \
  --why "What motivated it" \
  --where "Files or paths affected" \
  --learned "Gotchas, edge cases, things that surprised you" \
  --tags "tag1,tag2" \
  --topic-key "category/topic"
```

### TOPIC KEY RULES

- Use `{category}/{topic}` format: `architecture/auth-model`, `bug/n-plus-1-users`
- Reuse the same `--topic-key` to update an evolving topic instead of creating duplicates
- Different topics must not overwrite each other

### WHEN TO SEARCH

When the user asks to recall something — any variation of "remember", "recall", "what did we do":
```
cortex search "query" --project PROJECT
```

Also search PROACTIVELY when:
- Starting work on something that might have been done before
- The user mentions a topic you have no context on

### SESSION PROTOCOL

1. **On start**: Session is created automatically by the hook
2. **During work**: Use `--session SESSION` with `cortex save` to associate memories
3. **Before ending**: Save a comprehensive session summary:
```
cortex save \
  --title "Session summary" \
  --type decision \
  --project PROJECT \
  --session SESSION \
  --what "Goal: ... / Accomplished: ... / Next steps: ..." \
  --why "Session close protocol" \
  --where "files changed" \
  --learned "key discoveries"
```
4. Then: `cortex session end --id SESSION --summary "brief summary"`

### AFTER COMPACTION

If you see a message about compaction or context reset:
1. Save the compacted summary as a memory using `cortex save`
2. Run `cortex context --project PROJECT --session SESSION` to recover context
3. Only then continue working
