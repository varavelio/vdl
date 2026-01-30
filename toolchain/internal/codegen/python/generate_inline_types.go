package python

import (
	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/core/ir"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

type inlineTypeInfo struct {
	name   string
	doc    string
	fields []ir.Field
}

func extractInlineTypes(parentName string, tr ir.TypeRef) []inlineTypeInfo {
	var result []inlineTypeInfo

	switch tr.Kind {
	case ir.TypeKindObject:
		if tr.Object != nil {
			result = append(result, inlineTypeInfo{
				name:   parentName,
				doc:    "",
				fields: tr.Object.Fields,
			})
			for _, f := range tr.Object.Fields {
				childName := parentName + strutil.ToPascalCase(f.Name)
				result = append(result, extractInlineTypes(childName, f.Type)...)
			}
		}

	case ir.TypeKindArray:
		if tr.ArrayItem != nil {
			result = append(result, extractInlineTypes(parentName, *tr.ArrayItem)...)
		}

	case ir.TypeKindMap:
		if tr.MapValue != nil {
			result = append(result, extractInlineTypes(parentName, *tr.MapValue)...)
		}
	}

	return result
}

func extractAllInlineTypes(parentName string, fields []ir.Field) []inlineTypeInfo {
	var result []inlineTypeInfo
	for _, field := range fields {
		childName := parentName + strutil.ToPascalCase(field.Name)
		inlines := extractInlineTypes(childName, field.Type)
		result = append(result, inlines...)
	}
	return result
}

func renderInlineTypes(parentName string, fields []ir.Field) string {
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

func renderInlineType(name, doc string, fields []ir.Field) string {
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
