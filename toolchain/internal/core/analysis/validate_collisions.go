package analysis

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// SyntheticNameKind indicates the category of definition that generates synthetic names.
type SyntheticNameKind string

const (
	SyntheticNameKindEnum SyntheticNameKind = "enum"
	SyntheticNameKindProc SyntheticNameKind = "proc"
)

// ReservationRule defines a rule for reserving synthetic names.
// The rule specifies either a prefix or suffix that will be applied
// to a source definition's name to create reserved synthetic names.
type ReservationRule struct {
	Kind   SyntheticNameKind // The kind of source definition this rule applies to
	Prefix string            // Prefix to add (e.g., "is" -> "isColor")
	Suffix string            // Suffix to add (e.g., "List" -> "ColorList")
}

// Apply generates the synthetic name by applying this rule to a source name.
// Returns the synthetic name that would be generated.
func (r ReservationRule) Apply(sourceName string) string {
	return r.Prefix + sourceName + r.Suffix
}

// Description returns a human-readable description of what this rule generates.
func (r ReservationRule) Description() string {
	if r.Prefix != "" && r.Suffix != "" {
		return fmt.Sprintf("%s<Name>%s", r.Prefix, r.Suffix)
	}
	if r.Prefix != "" {
		return fmt.Sprintf("%s<Name>", r.Prefix)
	}
	return fmt.Sprintf("<Name>%s", r.Suffix)
}

// syntheticNameRules contains all the rules for generating synthetic names.
// This configuration can be extended in the future without modifying the validation logic.
var syntheticNameRules = []ReservationRule{
	// Enum rules
	{Kind: SyntheticNameKindEnum, Prefix: "is"},    // isColor - validation function
	{Kind: SyntheticNameKindEnum, Suffix: "List"},  // ColorList - array with all values
	{Kind: SyntheticNameKindEnum, Suffix: "Value"}, // ColorValue - value type alias

	// RPC/Proc rules
	{Kind: SyntheticNameKindProc, Suffix: "Input"},  // EchoInput - input type
	{Kind: SyntheticNameKindProc, Suffix: "Output"}, // EchoOutput - output type
}

// syntheticNameOrigin tracks the origin of a reserved synthetic name.
type syntheticNameOrigin struct {
	syntheticName  string            // The reserved synthetic name (e.g., "ColorList")
	sourceName     string            // The name of the definition that reserves it (e.g., "Color")
	sourceKind     SyntheticNameKind // The kind of the source (e.g., "enum")
	ruleDesc       string            // Description of the rule (e.g., "<Name>List")
	sourceFile     string            // File where the source is defined
	sourceCategory string            // Category for message (e.g., "enum", "procedure")
}

// validateCollisions checks that user-defined names don't collide with
// auto-generated synthetic names from other definitions.
//
// This validation protects users from generating code that will fail to compile
// in target languages due to duplicate identifiers.
func validateCollisions(symbols *symbolTable) []Diagnostic {
	var diagnostics []Diagnostic

	// Build index of all user-defined names with their location info
	userDefinedNames := buildUserDefinedNamesIndex(symbols)

	// Build index of all reserved synthetic names
	reservedSynthetic := buildReservedSyntheticIndex(symbols)

	// Check for collisions: user-defined names that match reserved synthetic names
	for name, origin := range reservedSynthetic {
		if userDef, exists := userDefinedNames[name]; exists {
			// Don't report collision if the synthetic name comes from the same definition
			// (e.g., don't report "ColorList" colliding with itself if user defines "ColorList")
			if userDef.name == origin.sourceName {
				continue
			}

			diagnostics = append(diagnostics, newDiagnostic(
				userDef.file,
				userDef.pos,
				userDef.endPos,
				CodeSyntheticNameCollision,
				formatSyntheticCollisionError(name, userDef.category, origin),
			))
		}
	}

	return diagnostics
}

// userDefinedName tracks information about a user-defined name.
type userDefinedName struct {
	name     string
	category string // "type", "enum", "pattern", "const"
	file     string
	pos      ast.Position
	endPos   ast.Position
}

// buildUserDefinedNamesIndex creates an index of all user-defined identifiers.
func buildUserDefinedNamesIndex(symbols *symbolTable) map[string]userDefinedName {
	index := make(map[string]userDefinedName)

	// Index types
	for name, sym := range symbols.types {
		index[name] = userDefinedName{
			name:     name,
			category: "type",
			file:     sym.File,
			pos:      sym.Pos,
			endPos:   sym.EndPos,
		}
	}

	// Index enums
	for name, sym := range symbols.enums {
		index[name] = userDefinedName{
			name:     name,
			category: "enum",
			file:     sym.File,
			pos:      sym.Pos,
			endPos:   sym.EndPos,
		}
	}

	// Index patterns
	for name, sym := range symbols.patterns {
		index[name] = userDefinedName{
			name:     name,
			category: "pattern",
			file:     sym.File,
			pos:      sym.Pos,
			endPos:   sym.EndPos,
		}
	}

	// Index consts (though less likely to collide due to UPPER_SNAKE_CASE)
	for name, sym := range symbols.consts {
		index[name] = userDefinedName{
			name:     name,
			category: "constant",
			file:     sym.File,
			pos:      sym.Pos,
			endPos:   sym.EndPos,
		}
	}

	return index
}

// buildReservedSyntheticIndex creates an index of all reserved synthetic names.
func buildReservedSyntheticIndex(symbols *symbolTable) map[string]syntheticNameOrigin {
	index := make(map[string]syntheticNameOrigin)

	// Apply enum rules to all enums
	for name, sym := range symbols.enums {
		for _, rule := range syntheticNameRules {
			if rule.Kind != SyntheticNameKindEnum {
				continue
			}
			syntheticName := rule.Apply(name)
			index[syntheticName] = syntheticNameOrigin{
				syntheticName:  syntheticName,
				sourceName:     name,
				sourceKind:     rule.Kind,
				ruleDesc:       rule.Description(),
				sourceFile:     sym.File,
				sourceCategory: "enum",
			}
		}
	}

	// Apply proc/stream rules to all RPCs
	for _, rpc := range symbols.rpcs {
		// Apply to procedures
		for procName, proc := range rpc.Procs {
			for _, rule := range syntheticNameRules {
				if rule.Kind != SyntheticNameKindProc {
					continue
				}
				syntheticName := rule.Apply(procName)
				index[syntheticName] = syntheticNameOrigin{
					syntheticName:  syntheticName,
					sourceName:     procName,
					sourceKind:     rule.Kind,
					ruleDesc:       rule.Description(),
					sourceFile:     proc.File,
					sourceCategory: "procedure",
				}
			}
		}

		// Apply to streams (same rules as procs)
		for streamName, stream := range rpc.Streams {
			for _, rule := range syntheticNameRules {
				if rule.Kind != SyntheticNameKindProc {
					continue
				}
				syntheticName := rule.Apply(streamName)
				index[syntheticName] = syntheticNameOrigin{
					syntheticName:  syntheticName,
					sourceName:     streamName,
					sourceKind:     rule.Kind,
					ruleDesc:       rule.Description(),
					sourceFile:     stream.File,
					sourceCategory: "stream",
				}
			}
		}
	}

	return index
}

// formatSyntheticCollisionError creates a descriptive error message for synthetic name collisions.
func formatSyntheticCollisionError(name, userCategory string, origin syntheticNameOrigin) string {
	return fmt.Sprintf(
		"%s %q is reserved because it is auto-generated from %s %q (pattern: %s). Please rename your definition.",
		userCategory,
		name,
		origin.sourceCategory,
		origin.sourceName,
		origin.ruleDesc,
	)
}
