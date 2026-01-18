package ir

import (
	"strings"
)

// normalizeDoc processes a docstring for the IR.
// It strips common leading indentation and trims whitespace.
// Returns empty string if input is nil or empty.
func normalizeDoc(raw *string) string {
	if raw == nil {
		return ""
	}

	s := *raw
	if s == "" {
		return ""
	}

	// Split into lines
	lines := strings.Split(s, "\n")

	// Find minimum indentation (ignoring empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := 0
		for _, ch := range line {
			switch ch {
			case ' ':
				indent++
			case '\t':
				indent += 4 // Treat tab as 4 spaces
			default:
				goto done
			}
		}
	done:
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// If no indentation found, just trim
	if minIndent <= 0 {
		return strings.TrimSpace(s)
	}

	// Remove common indentation from each line
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			result = append(result, "")
			continue
		}

		// Count leading whitespace and remove up to minIndent
		removed := 0
		idx := 0
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
				goto done2
			}
		}
	done2:
		result = append(result, line[idx:])
	}

	// Join and trim
	return strings.TrimSpace(strings.Join(result, "\n"))
}
