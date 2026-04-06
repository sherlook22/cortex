package sqlite

import "testing"

func TestSanitizeFTS(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple words", input: "fix auth bug", want: `"fix" "auth" "bug"`},
		{name: "single word", input: "auth", want: `"auth"`},
		{name: "empty string", input: "", want: ""},
		{name: "only spaces", input: "   ", want: ""},
		{name: "removes existing quotes", input: `"auth" "bug"`, want: `"auth" "bug"`},
		{name: "handles special chars", input: "fix:auth+bug", want: `"fix:auth+bug"`},
		{name: "collapses whitespace", input: "fix   auth   bug", want: `"fix" "auth" "bug"`},
		{name: "strips embedded quotes", input: `he"llo wo"rld`, want: `"hello" "world"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFTS(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFTS(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
