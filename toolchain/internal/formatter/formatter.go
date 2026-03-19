package formatter

import (
	"strings"
)

// Format formats VDL content according to the formatting guide.
//
// The formatter is lexer-based so it can preserve comment content and keep
// formatting stable without depending on the semantic parser tree.
func Format(filename, content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	// Check if the file starts with // format:off or //format:off
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "// format:off") || strings.HasPrefix(trimmed, "//format:off") {
		return content, nil
	}

	return formatLexerBased(filename, content)
}
