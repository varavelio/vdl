package strutil

import (
	"strings"
)

// NormalizeIndent removes the common leading indentation from every line in text.
// It computes the minimum indentation across all non-empty lines (counting tabs as 4 spaces)
// and strips that amount of visual indentation from each line.
// It preserves relative indentation and vertical whitespace.
func NormalizeIndent(text string) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")

	// Find minimum indentation (ignoring empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := 0
	countLoop:
		for _, ch := range line {
			switch ch {
			case ' ':
				indent++
			case '\t':
				indent += 4 // Treat tab as 4 spaces
			default:
				break countLoop
			}
		}

		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// If no indentation found, default to 0 so we still process lines
	// (e.g. to clean up whitespace-only lines)
	if minIndent == -1 {
		minIndent = 0
	}

	// Remove common indentation from each line
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		// If line is empty or purely whitespace, we replace it with empty string
		// to avoid trailing whitespace, which is standard behavior for dedent.
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
			continue
		}

		// Count leading whitespace and remove up to minIndent
		removed := 0
		idx := 0
	stripLoop:
		for i, ch := range line {
			if removed >= minIndent {
				break
			}
			switch ch {
			case ' ':
				removed++
				idx = i + 1
			case '\t':
				removed += 4
				idx = i + 1
			default:
				break stripLoop
			}
		}
		result = append(result, line[idx:])
	}

	return strings.Join(result, "\n")
}
