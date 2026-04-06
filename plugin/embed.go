// Package plugin provides embedded plugin assets for distribution.
package plugin

import "embed"

//go:embed all:claude-code
var ClaudeCodeFS embed.FS

//go:embed opencode/cortex.ts
var OpenCodePlugin []byte
