package analysis

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// validatePatterns validates all pattern declarations:
// - Placeholder syntax must be valid {name}
// - Placeholder names must be valid identifiers
func validatePatterns(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	for _, pattern := range symbols.patterns {
		diagnostics = append(diagnostics, validatePattern(pattern)...)
	}

	return diagnostics
}

// placeholderRegex matches valid placeholders: {identifier}
var placeholderRegex = regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// invalidPlaceholderRegex matches malformed placeholders
var invalidPlaceholderRegex = regexp.MustCompile(`\{[^}]*\}|\{[^}]*$`)

// validatePattern validates a single pattern declaration.
func validatePattern(pattern *PatternSymbol) []Diagnostic {
	var diagnostics []Diagnostic
	template := pattern.Template

	// Find all valid placeholders
	validMatches := placeholderRegex.FindAllStringSubmatchIndex(template, -1)
	validRanges := make(map[int]int) // start -> end
	for _, match := range validMatches {
		validRanges[match[0]] = match[1]
	}

	// Find all placeholder-like patterns (including malformed ones)
	allMatches := invalidPlaceholderRegex.FindAllStringIndex(template, -1)

	for _, match := range allMatches {
		start, end := match[0], match[1]
		placeholder := template[start:end]

		// Check if this is a valid placeholder
		if _, isValid := validRanges[start]; isValid {
			continue
		}

		// This is a malformed placeholder
		if !strings.HasSuffix(placeholder, "}") {
			diagnostics = append(diagnostics, newDiagnostic(
				pattern.File,
				pattern.Pos,
				pattern.EndPos,
				CodePatternInvalidSyntax,
				fmt.Sprintf("unclosed placeholder in pattern %q: %q", pattern.Name, placeholder),
			))
		} else {
			// Has braces but content is not a valid identifier
			content := strings.TrimPrefix(strings.TrimSuffix(placeholder, "}"), "{")
			diagnostics = append(diagnostics, newDiagnostic(
				pattern.File,
				pattern.Pos,
				pattern.EndPos,
				CodePatternInvalidPlaceholder,
				fmt.Sprintf("invalid placeholder name %q in pattern %q: must be a valid identifier", content, pattern.Name),
			))
		}
	}

	// Check for unclosed braces at the end
	if strings.Count(template, "{") != strings.Count(template, "}") {
		diagnostics = append(diagnostics, newDiagnostic(
			pattern.File,
			pattern.Pos,
			pattern.EndPos,
			CodePatternInvalidSyntax,
			fmt.Sprintf("mismatched braces in pattern %q", pattern.Name),
		))
	}

	return diagnostics
}

// extractPlaceholders extracts all placeholder names from a template string.
func extractPlaceholders(template string) []string {
	matches := placeholderRegex.FindAllStringSubmatch(template, -1)
	placeholders := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			placeholders = append(placeholders, match[1])
		}
	}
	return placeholders
}

// buildPatternSymbol creates a PatternSymbol from an AST PatternDecl.
func buildPatternSymbol(decl *ast.PatternDecl, file string) *PatternSymbol {
	var docstring *string
	if decl.Docstring != nil {
		s := string(decl.Docstring.Value)
		docstring = &s
	}

	var deprecated *DeprecationInfo
	if decl.Deprecated != nil {
		msg := ""
		if decl.Deprecated.Message != nil {
			msg = string(*decl.Deprecated.Message)
		}
		deprecated = &DeprecationInfo{Message: msg}
	}

	template := string(decl.Value)
	return &PatternSymbol{
		Symbol: Symbol{
			Name:       decl.Name,
			File:       file,
			Pos:        decl.Pos,
			EndPos:     decl.EndPos,
			Docstring:  docstring,
			Deprecated: deprecated,
		},
		AST:          decl,
		Template:     template,
		Placeholders: extractPlaceholders(template),
	}
}
