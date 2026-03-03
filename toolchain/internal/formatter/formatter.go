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

	// Check if the file starts with // fmt:off or //fmt:off
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "// fmt:off") || strings.HasPrefix(trimmed, "//fmt:off") {
		return content, nil
	}

	return formatLexerBased(filename, content)
}
