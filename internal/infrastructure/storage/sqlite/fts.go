package sqlite

import "strings"

// sanitizeFTS wraps each word in double quotes to prevent FTS5 syntax errors
// from special characters. All terms are implicit AND.
// Example: "fix auth bug" -> `"fix" "auth" "bug"`
func sanitizeFTS(query string) string {
	words := strings.Fields(query)
	if len(words) == 0 {
		return ""
	}

	quoted := make([]string, len(words))
	for i, w := range words {
		w = strings.Trim(w, `"`)
		w = strings.ReplaceAll(w, `"`, ``)
		if w != "" {
			quoted[i] = `"` + w + `"`
		}
	}

	return strings.Join(quoted, " ")
}
