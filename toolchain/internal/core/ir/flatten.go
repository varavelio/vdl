package ir

import (
	"strconv"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
)

func flattenTypeFields(
	typ *analysis.TypeSymbol,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	resolver *valueResolver,
) []irtypes.Field {
	if typ == nil {
		return nil
	}
	return flattenFieldsWithSpreads(typ.Fields, typ.Spreads, types, enums, resolver, map[string]bool{typ.Name: true})
}

func flattenFieldsWithSpreads(
	fields []*analysis.FieldSymbol,
	spreads []*analysis.SpreadRef,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	resolver *valueResolver,
	visiting map[string]bool,
) []irtypes.Field {
	result := make([]irtypes.Field, 0, len(fields))

	for _, spread := range spreads {
		if spread == nil || spread.Member != nil {
			continue
		}
		spreadType := types[spread.Name]
		if spreadType == nil || visiting[spreadType.Name] {
			continue
		}

		nextVisiting := cloneVisited(visiting)
		nextVisiting[spreadType.Name] = true
		spreadFields := flattenFieldsWithSpreads(spreadType.Fields, spreadType.Spreads, types, enums, resolver, nextVisiting)
		result = append(result, spreadFields...)
	}

	for _, field := range fields {
		result = append(result, convertField(field, types, enums, resolver))
	}

	return result
}

func flattenInlineObjectFields(
	obj *analysis.InlineObject,
	types map[string]*analysis.TypeSymbol,
	enums map[string]*analysis.EnumSymbol,
	resolver *valueResolver,
) *[]irtypes.Field {
	if obj == nil {
		return nil
	}
	fields := flattenFieldsWithSpreads(obj.Fields, obj.Spreads, types, enums, resolver, map[string]bool{})
	return &fields
}

func expandEnumMembers(
	enum *analysis.EnumSymbol,
	enums map[string]*analysis.EnumSymbol,
	visiting map[string]bool,
) []*analysis.EnumMemberSymbol {
	if enum == nil {
		return nil
	}
	if visiting[enum.Name] {
		return nil
	}

	visiting[enum.Name] = true
	defer delete(visiting, enum.Name)

	members := make([]*analysis.EnumMemberSymbol, 0, len(enum.Members))
	for _, spread := range enum.Spreads {
		if spread == nil || spread.Member != nil {
			continue
		}
		spreadEnum := enums[spread.Name]
		if spreadEnum == nil {
			continue
		}
		members = append(members, expandEnumMembers(spreadEnum, enums, visiting)...)
	}

	members = append(members, enum.Members...)
	return members
}

func lookupEnumMemberValue(
	enums map[string]*analysis.EnumSymbol,
	enumName, memberName string,
) (irtypes.Value, bool) {
	enum := enums[enumName]
	if enum == nil {
		return irtypes.Value{}, false
	}

	members := expandEnumMembers(enum, enums, map[string]bool{})
	for _, m := range members {
		if m.Name != memberName {
			continue
		}

		if enum.ValueType == analysis.EnumValueTypeInt {
			n, err := strconv.ParseInt(m.Value, 10, 64)
			if err != nil {
				return irtypes.Value{}, false
			}
			return irtypes.Value{
				Kind:     irtypes.ValueKindInt,
				IntValue: irtypes.Ptr(n),
			}, true
		}

		return irtypes.Value{
			Kind:        irtypes.ValueKindString,
			StringValue: irtypes.Ptr(m.Value),
		}, true
	}

	return irtypes.Value{}, false
}

func cloneVisited(src map[string]bool) map[string]bool {
	dst := make(map[string]bool, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
