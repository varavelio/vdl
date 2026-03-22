package formatter

import (
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func formatTypeName(name string) string {
	if isPrimitiveName(name) {
		return name
	}
	return strutil.ToPascalCase(name)
}

func isPrimitiveName(name string) bool {
	switch name {
	case "string", "int", "float", "bool", "datetime":
		return true
	default:
		return false
	}
}

func renderFieldType(ft fieldTypeNode) string {
	base := ""
	switch {
	case ft.Named != nil:
		base = formatTypeName(*ft.Named)
	case ft.Map != nil:
		base = "map[" + renderFieldType(*ft.Map) + "]"
	case ft.Obj != nil:
		base = renderObjectType(*ft.Obj)
	}
	for range ft.Dims {
		base += "[]"
	}
	return base
}

func renderObjectType(obj objectTypeNode) string {
	if len(obj.Members) == 0 {
		return "{}"
	}

	output := newFormatterOutput()
	output.Line("{")
	output.Block(func() {
		for i, member := range obj.Members {
			if i > 0 && hasTypeMemberBlankLine(obj.Members[i-1], member) {
				blankLine(output)
			}
			printTypeMember(output, member)
		}
	})
	output.Line("}")

	return strings.TrimSuffix(output.String(), "\n")
}

func renderLiteral(l literalNode, ctx literalRenderCtx) string {
	if l.Obj != nil {
		return renderObjectLiteral(*l.Obj, ctx)
	}
	if l.Array != nil {
		return renderArrayLiteral(*l.Array, ctx)
	}
	if l.Scalar != nil {
		return renderScalar(*l.Scalar, ctx)
	}
	return ""
}

func renderObjectLiteral(o objectLiteralNode, ctx literalRenderCtx) string {
	if len(o.Entries) == 0 {
		return "{}"
	}
	if !ctx.forceObjectMultiline && len(o.Entries) == 1 && o.Entries[0].Spread == nil && o.Entries[0].Comment == nil && o.Entries[0].Trailing == nil {
		entry := o.Entries[0]
		if entry.Value != nil && entry.Value.Scalar != nil {
			return "{ " + strutil.ToCamelCase(entry.Key) + " " + renderLiteral(*entry.Value, ctx) + " }"
		}
	}

	output := newFormatterOutput()
	output.Line("{")
	output.Block(func() {
		for i, entry := range o.Entries {
			if i > 0 && hasSourceBlankLine(o.Entries[i-1], entry) {
				blankLine(output)
			}
			switch {
			case entry.Comment != nil:
				output.Line(entry.Comment.Text)
			case entry.Spread != nil:
				lineWithTrailing(output, "..."+renderReference(*entry.Spread, ctx.spreadRef), entry.Trailing)
			default:
				key := strutil.ToCamelCase(entry.Key)
				value := renderLiteral(*entry.Value, ctx)
				writeRenderedValue(output, key+" ", value, entry.Trailing)
			}
		}
	})
	output.Line("}")

	return strings.TrimSuffix(output.String(), "\n")
}

func renderArrayLiteral(a arrayLiteralNode, ctx literalRenderCtx) string {
	if len(a.Elements) == 0 {
		return "[]"
	}

	parts := make([]string, 0, len(a.Elements))
	hasComments := false
	hasTrailingComments := false
	hasMultiline := false
	hasCompoundElements := false
	for _, element := range a.Elements {
		if element.Comment != nil {
			hasComments = true
			parts = append(parts, element.Comment.Text)
			continue
		}
		hasCompoundElements = hasCompoundElements || (element.Value != nil && (element.Value.Obj != nil || element.Value.Array != nil))
		rendered := renderLiteral(*element.Value, ctx)
		hasMultiline = hasMultiline || strings.Contains(rendered, "\n")
		hasTrailingComments = hasTrailingComments || element.Trailing != nil
		parts = append(parts, rendered)
	}

	shouldMultiline := hasComments || hasTrailingComments || hasMultiline
	if ctx.forceCompoundArrayMultiline && hasCompoundElements {
		shouldMultiline = true
	}
	if ctx.respectArrayMultilineIntent && a.MultilineIntent {
		shouldMultiline = true
	}
	if !shouldMultiline {
		return "[" + strings.Join(parts, " ") + "]"
	}

	output := newFormatterOutput()
	output.Line("[")
	output.Block(func() {
		for i, element := range a.Elements {
			if i > 0 && hasSourceBlankLine(a.Elements[i-1], element) {
				blankLine(output)
			}
			switch {
			case element.Comment != nil:
				output.Line(element.Comment.Text)
			case element.Value != nil:
				writeRenderedValue(output, "", renderLiteral(*element.Value, ctx), element.Trailing)
			}
		}
	})
	output.Line("]")

	return strings.TrimSuffix(output.String(), "\n")
}

type lineSpan interface {
	startLine() int
	endLine() int
}

func hasSourceBlankLine(prev, curr lineSpan) bool {
	if prev == nil || curr == nil {
		return false
	}
	return curr.startLine()-prev.endLine() > 1
}

func renderScalar(s scalarLiteralNode, ctx literalRenderCtx) string {
	switch {
	case s.Str != nil:
		return `"` + strutil.EscapeQuotes(*s.Str) + `"`
	case s.Float != nil:
		return *s.Float
	case s.Int != nil:
		return *s.Int
	case s.True:
		return "true"
	case s.False:
		return "false"
	case s.Ref != nil && s.Ref.Member != nil:
		return renderReference(*s.Ref, ctx.enumMemberRef)
	case s.Ref != nil:
		return renderReference(*s.Ref, ctx.scalarRef)
	default:
		return ""
	}
}

func renderReference(r referenceNode, c refCase) string {
	name := r.Name
	switch c {
	case refTypeDecl, refEnumDecl:
		name = strutil.ToPascalCase(name)
	case refConstDecl:
		name = strutil.ToCamelCase(name)
	}
	if r.Member != nil {
		return strutil.ToPascalCase(name) + "." + strutil.ToPascalCase(*r.Member)
	}
	return name
}
