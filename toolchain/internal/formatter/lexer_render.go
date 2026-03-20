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
	for i := 0; i < ft.Dims; i++ {
		base += "[]"
	}
	return base
}

func renderObjectType(obj objectTypeNode) string {
	if len(obj.Members) == 0 {
		return "{}"
	}
	w := newFmtWriter()
	w.line("{")
	w.indent++
	for i, m := range obj.Members {
		if i > 0 && hasTypeMemberBlankLine(obj.Members[i-1], m) {
			w.blank()
		}
		printTypeMember(w, m)
	}
	w.indent--
	w.line("}")
	return strings.TrimSuffix(w.String(), "\n")
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
		e := o.Entries[0]
		if e.Value != nil && e.Value.Scalar != nil {
			return "{ " + strutil.ToCamelCase(e.Key) + " " + renderLiteral(*e.Value, ctx) + " }"
		}
	}

	w := newFmtWriter()
	w.line("{")
	w.indent++
	for i, e := range o.Entries {
		if i > 0 && hasSourceBlankLine(o.Entries[i-1], e) {
			w.blank()
		}
		switch {
		case e.Comment != nil:
			w.line(e.Comment.Text)
		case e.Spread != nil:
			w.lineWithTrailing("..."+renderReference(*e.Spread, ctx.spreadRef), e.Trailing)
		default:
			key := strutil.ToCamelCase(e.Key)
			val := renderLiteral(*e.Value, ctx)
			writeRenderedValue(w, key+" ", val, e.Trailing)
		}
	}
	w.indent--
	w.line("}")
	return strings.TrimSuffix(w.String(), "\n")
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
	for _, e := range a.Elements {
		if e.Comment != nil {
			hasComments = true
			parts = append(parts, e.Comment.Text)
			continue
		}
		hasCompoundElements = hasCompoundElements || (e.Value != nil && (e.Value.Obj != nil || e.Value.Array != nil))
		rendered := renderLiteral(*e.Value, ctx)
		hasMultiline = hasMultiline || strings.Contains(rendered, "\n")
		hasTrailingComments = hasTrailingComments || e.Trailing != nil
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

	w := newFmtWriter()
	w.line("[")
	w.indent++
	for i, e := range a.Elements {
		if i > 0 && hasSourceBlankLine(a.Elements[i-1], e) {
			w.blank()
		}
		switch {
		case e.Comment != nil:
			w.line(e.Comment.Text)
		case e.Value != nil:
			writeRenderedValue(w, "", renderLiteral(*e.Value, ctx), e.Trailing)
		}
	}
	w.indent--
	w.line("]")
	return strings.TrimSuffix(w.String(), "\n")
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

func writeRenderedValue(w *fmtWriter, prefix, value string, trailing *commentNode) {
	if !strings.Contains(value, "\n") {
		w.lineWithTrailing(prefix+value, trailing)
		return
	}

	lines := strings.Split(value, "\n")
	w.line(prefix + lines[0])
	for i := 1; i < len(lines)-1; i++ {
		if lines[i] == "" {
			w.blank()
			continue
		}
		w.line(lines[i])
	}
	last := lines[len(lines)-1]
	if trailing != nil {
		w.line(last + " " + trailing.Text)
		return
	}
	w.line(last)
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
