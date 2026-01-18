package ir

import (
	"github.com/uforg/uforpc/urpc/internal/core/analysis"
)

// flattenFields expands all spreads into a flat list of fields.
// Spread fields come first (in declaration order), then local fields.
// This function recursively resolves nested spreads.
func flattenFields(
	fields []*analysis.FieldSymbol,
	spreads []*analysis.SpreadRef,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) []Field {
	result := make([]Field, 0)

	// First, add fields from spreads (in order)
	for _, spread := range spreads {
		spreadType, ok := types[spread.TypeName]
		if !ok {
			// Type not found - skip (should have been caught by analysis)
			continue
		}

		// Recursively flatten the spread type's fields
		spreadFields := flattenFields(spreadType.Fields, spreadType.Spreads, types, enums)
		result = append(result, spreadFields...)
	}

	// Then, add local fields
	for _, field := range fields {
		result = append(result, convertField(field, types, enums))
	}

	return result
}

// flattenBlockFields expands spreads in a proc/stream input/output block.
func flattenBlockFields(
	block *analysis.BlockSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) []Field {
	if block == nil {
		return []Field{}
	}
	return flattenFields(block.Fields, block.Spreads, types, enums)
}

// flattenInlineObject expands spreads in an inline object type.
func flattenInlineObject(
	obj *analysis.InlineObject,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) *InlineObject {
	if obj == nil {
		return nil
	}
	return &InlineObject{
		Fields: flattenFields(obj.Fields, obj.Spreads, types, enums),
	}
}
