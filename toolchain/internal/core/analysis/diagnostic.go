// Package analysis provides semantic analysis for VDL schemas.
// It validates the meaning of the code, resolves imports, checks types,
// and produces a unified Program with all symbols merged.
package analysis

import (
	"fmt"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// Error codes for diagnostics.
// Resolution errors (E00x)
const (
	CodeFileNotFound          = "E001"
	CodeCircularInclude       = "E002"
	CodeDocstringFileNotFound = "E003"
	CodeFileReadError         = "E004"
	CodeParseError            = "E005"
)

// Naming errors (E10x)
const (
	CodeNotPascalCase       = "E101"
	CodeNotCamelCase        = "E102"
	CodeEnumMemberNotPascal = "E103"
)

// Type reference errors (E20x)
const (
	CodeTypeNotDeclared      = "E201"
	CodeSpreadTypeNotFound   = "E202"
	CodeSpreadFieldConflict  = "E203"
	CodeSpreadCycle          = "E204"
	CodeInvalidReference     = "E205"
	CodeConstSpreadNotObject = "E206"
	CodeConstArrayMixedTypes = "E207"
)

// Enum errors (E30x)
const (
	CodeEnumMixedTypes     = "E301"
	CodeEnumIntNeedsValues = "E302"
	CodeEnumDuplicateValue = "E303"
	CodeEnumDuplicateName  = "E304"
	CodeEnumMemberNotFound = "E305"
)

// Cycle errors (E60x)
const (
	CodeCircularTypeDependency = "E601"
)

// Structure errors (E70x)
const (
	CodeDuplicateField = "E701"
)

// Global uniqueness errors (E80x)
const (
	CodeDuplicateType  = "E801"
	CodeDuplicateEnum  = "E802"
	CodeDuplicateConst = "E803"
	CodeDuplicateName  = "E804" // Cross-category name collision
)

// Diagnostic represents an error found during semantic analysis.
// It provides precise location information for IDE/LSP integration.
type Diagnostic struct {
	File    string       // The file where the error occurred
	Pos     ast.Position // Start position of the error
	EndPos  ast.Position // End position of the error
	Code    string       // Error code (e.g., "E001")
	Message string       // Human-readable error message
}

// String returns a formatted string representation of the diagnostic.
// Format: "file:line:column: error[CODE]: message"
func (d Diagnostic) String() string {
	return fmt.Sprintf(
		"%s:%d:%d: error[%s]: %s",
		d.File, d.Pos.Line, d.Pos.Column, d.Code, d.Message,
	)
}

// Error implements the error interface, returning the same as String().
func (d Diagnostic) Error() string {
	return d.String()
}

// newDiagnostic creates a new Diagnostic with the given parameters.
func newDiagnostic(file string, pos, endPos ast.Position, code, message string) Diagnostic {
	return Diagnostic{
		File:    file,
		Pos:     pos,
		EndPos:  endPos,
		Code:    code,
		Message: message,
	}
}

// newDiagnosticFromPositions creates a Diagnostic from ast.Positions.
func newDiagnosticFromPositions(positions ast.Positions, code, message string) Diagnostic {
	return Diagnostic{
		File:    positions.Pos.Filename,
		Pos:     positions.Pos,
		EndPos:  positions.EndPos,
		Code:    code,
		Message: message,
	}
}

// formatSuggestions formats a slice of suggestions for display in error messages.
// Returns strings like: "Foo", "Foo or Bar", "Foo, Bar, or Baz"
func formatSuggestions(suggestions []string) string {
	switch len(suggestions) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("%q", suggestions[0])
	case 2:
		return fmt.Sprintf("%q or %q", suggestions[0], suggestions[1])
	default:
		var b strings.Builder
		for i, s := range suggestions {
			if i == len(suggestions)-1 {
				fmt.Fprintf(&b, "or %q", s)
			} else {
				fmt.Fprintf(&b, "%q, ", s)
			}
		}
		return b.String()
	}
}
