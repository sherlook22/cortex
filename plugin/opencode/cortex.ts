import { execSync, spawnSync } from "child_process";
import type { Plugin } from "@opencode-ai/opencode";

// --- Configuration ---

const CORTEX_BIN = process.env.CORTEX_BIN || "cortex";

// --- Helper ---

function cortex(args: string): string {
  try {
    return execSync(`${CORTEX_BIN} ${args}`, {
      encoding: "utf-8",
      timeout: 10000,
    }).trim();
  } catch {
    return "";
  }
}

function cortexAsync(args: string, input?: string): void {
  try {
    spawnSync(CORTEX_BIN, args.split(" "), {
      input,
      encoding: "utf-8",
      timeout: 10000,
    });
  } catch {
    // Silent failure — never crash the host.
  }
}

// --- State ---

const knownSessions = new Set<string>();
const toolCounts = new Map<string, number>();

function projectFromDir(dir: string): string {
  return dir.split("/").pop() || "unknown";
}

function ensureSession(sessionId: string, project: string, directory: string): void {
  if (knownSessions.has(sessionId)) return;
  cortex(`session start --id ${sessionId} --project ${project} --directory ${directory}`);
  knownSessions.add(sessionId);
}

// --- Memory Protocol ---

const MEMORY_INSTRUCTIONS = `
## Cortex Persistent Memory — Protocol

You have access to cortex, a persistent memory CLI tool.

### WHEN TO SAVE (mandatory)
Run \`cortex save\` IMMEDIATELY after: bug fixes, architecture decisions, discoveries, config changes, patterns, preferences.

Format:
  cortex save --title "Verb + what" --type <bugfix|decision|architecture|discovery|pattern|config> --project PROJECT --session SESSION --what "..." --why "..." --where "..." --learned "..." --tags "..." --topic-key "category/topic"

### WHEN TO SEARCH
Run \`cortex search "query" --project PROJECT\` when the user asks to recall something, or proactively when starting related work.

### SESSION CLOSE (mandatory)
Before ending, save a session summary with cortex save, then run:
  cortex session end --id SESSION --summary "brief summary"

### AFTER COMPACTION
1. Save compacted summary as a memory
2. Run \`cortex context --project PROJECT --session SESSION\` to recover
3. Continue working
`;

// --- Plugin ---

const plugin: Plugin = {
  name: "cortex",
  version: "0.1.0",

  event(event) {
    if (event.type === "session.created") {
      const session = event.properties.session;
      const project = projectFromDir(session.directory || process.cwd());
      ensureSession(session.id, project, session.directory || process.cwd());
    }

    if (event.type === "session.deleted") {
      const session = event.properties.session;
      knownSessions.delete(session.id);
      toolCounts.delete(session.id);
    }
  },

  hooks: {
    "tool.execute.after"(event) {
      const { tool, session, output } = event.properties;

      // Skip cortex's own tools.
      if (tool.name?.startsWith("cortex")) return;

      // Count tool calls per session.
      const count = (toolCounts.get(session.id) || 0) + 1;
      toolCounts.set(session.id, count);

      // Passive capture for Task tool output.
      if (tool.name === "Task" && output && output.length > 50) {
        const project = projectFromDir(session.directory || process.cwd());
        ensureSession(session.id, project, session.directory || process.cwd());
        cortexAsync(
          `capture --project ${project} --session ${session.id} --source subagent`,
          output
        );
      }
    },

    "experimental.chat.system.transform"(event) {
      const messages = event.properties.messages;
      if (messages.length > 0) {
        // Append to last system message to avoid multiple system blocks.
        const last = messages[messages.length - 1];
        last.content += "\n\n" + MEMORY_INSTRUCTIONS;
      } else {
        messages.push({ role: "system", content: MEMORY_INSTRUCTIONS });
      }
      return messages;
    },

    "experimental.session.compacting"(event) {
      const { session } = event.properties;
      const project = projectFromDir(session.directory || process.cwd());

      ensureSession(session.id, project, session.directory || process.cwd());

      // Fetch previous context.
      const context = cortex(`context --project ${project} --session ${session.id}`);

      let injection = "\n\n## Cortex Memory — Compaction Recovery\n";
      if (context) {
        injection += "\n" + context;
      }
      injection += `
CRITICAL INSTRUCTION POST-COMPACTION:
FIRST ACTION REQUIRED: Save the compacted summary above as a memory using:
  cortex save --title "Session summary (compacted)" --type decision --project ${project} --session ${session.id} --what "..." --why "Compaction recovery" --where "..." --learned "..."

Then run \`cortex context --project ${project} --session ${session.id}\` and continue working.
`;

      return injection;
    },
  },
};

export default plugin;
