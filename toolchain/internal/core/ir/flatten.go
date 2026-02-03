package ir

import (
	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

// flattenFields expands all spreads into a flat list of fields.
// Spread fields come first (in declaration order), then local fields.
// This function recursively resolves nested spreads.
func flattenFields(
	fields []*analysis.FieldSymbol,
	spreads []*analysis.SpreadRef,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) []irtypes.Field {
	result := make([]irtypes.Field, 0)

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
) []irtypes.Field {
	if block == nil {
		return []irtypes.Field{}
	}
	return flattenFields(block.Fields, block.Spreads, types, enums)
}

// flattenInlineObjectFields expands spreads in an inline object type and returns the fields.
func flattenInlineObjectFields(
	obj *analysis.InlineObject,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
) *[]irtypes.Field {
	if obj == nil {
		return nil
	}
	fields := flattenFields(obj.Fields, obj.Spreads, types, enums)
	return &fields
}
