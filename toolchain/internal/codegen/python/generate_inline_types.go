package python

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type inlineTypeInfo struct {
	name   string
	doc    string
	fields []irtypes.Field
}

func extractInlineTypes(parentName string, tr irtypes.TypeRef) []inlineTypeInfo {
	var result []inlineTypeInfo

	switch tr.Kind {
	case irtypes.TypeKindObject:
		if tr.ObjectFields != nil {
			result = append(result, inlineTypeInfo{
				name:   parentName,
				doc:    "",
				fields: *tr.ObjectFields,
			})
			for _, f := range *tr.ObjectFields {
				childName := parentName + strutil.ToPascalCase(f.Name)
				result = append(result, extractInlineTypes(childName, f.TypeRef)...)
			}
		}

	case irtypes.TypeKindArray:
		if tr.ArrayType != nil {
			result = append(result, extractInlineTypes(parentName, *tr.ArrayType)...)
		}

	case irtypes.TypeKindMap:
		if tr.MapType != nil {
			result = append(result, extractInlineTypes(parentName, *tr.MapType)...)
		}
	}

	return result
}

func extractAllInlineTypes(parentName string, fields []irtypes.Field) []inlineTypeInfo {
	var result []inlineTypeInfo
	for _, field := range fields {
		childName := parentName + strutil.ToPascalCase(field.Name)
		inlines := extractInlineTypes(childName, field.TypeRef)
		result = append(result, inlines...)
	}
	return result
}

func renderInlineTypes(parentName string, fields []irtypes.Field) string {
	inlineTypes := extractAllInlineTypes(parentName, fields)
	if len(inlineTypes) == 0 {
		return ""
	}

	g := gen.New()
	for _, inlineType := range inlineTypes {
		g.Raw(renderInlineType(inlineType.name, inlineType.doc, inlineType.fields))
		g.Break()
	}
	return g.String()
}

func renderInlineType(name, doc string, fields []irtypes.Field) string {
	if len(fields) == 0 {
		return ""
	}
	g := gen.New()
	if doc != "" {
		g.Linef("\"\"\"%s\"\"\"", doc)
	}
	g.Raw(GenerateDataclass(name, doc, fields))
	return g.String()
}
