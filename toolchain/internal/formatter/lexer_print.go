package formatter

import (
	"strings"

	"github.com/varavelio/gen"
	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func printDocument(output *gen.Generator, d *docNode) {
	for i, item := range d.Items {
		if i > 0 && hasTopLevelBlankLine(d.Items[i-1], item) {
			blankLine(output)
		}
		printTopNode(output, item)
	}
}

func hasTopLevelBlankLine(prev, curr node) bool {
	if prev == nil || curr == nil {
		return false
	}
	return topNodeStartLine(curr)-topNodeEndLine(prev) > 1
}

func printTopNode(output *gen.Generator, n node) {
	switch t := n.(type) {
	case *commentNode:
		output.Line(t.Text)
	case *docstringNode:
		printDocstring(output, t.Raw)
	case *includeNode:
		if t.Doc != nil {
			printDocstring(output, t.Doc.Raw)
		}
		for _, a := range t.Ann {
			printAnnotation(output, a)
		}
		lineWithTrailing(output, `include "`+strutil.EscapeQuotes(t.Path)+`"`, t.Trailing)
	case *typeNode:
		printTypeDecl(output, t)
	case *constNode:
		printConstDecl(output, t)
	case *enumNode:
		printEnumDecl(output, t)
	}
}

func printTypeDecl(output *gen.Generator, t *typeNode) {
	if t.Doc != nil {
		printDocstring(output, t.Doc.Raw)
	}
	for _, a := range t.Ann {
		printAnnotation(output, a)
	}

	name := strutil.ToPascalCase(t.Name)
	if t.Type.Obj == nil {
		writeRenderedValue(output, "type "+name+" ", renderFieldType(t.Type), t.Trailing)
		return
	}
	if len(t.Type.Obj.Members) == 0 {
		lineWithTrailing(output, "type "+name+" {}"+strings.Repeat("[]", t.Type.Dims), t.Trailing)
		return
	}

	output.Line("type " + name + " {")
	output.Block(func() {
		for i, m := range t.Type.Obj.Members {
			if i > 0 && hasTypeMemberBlankLine(t.Type.Obj.Members[i-1], m) {
				blankLine(output)
			}
			printTypeMember(output, m)
		}
	})
	lineWithTrailing(output, "}"+strings.Repeat("[]", t.Type.Dims), t.Trailing)
}

func printConstDecl(output *gen.Generator, t *constNode) {
	if t.Doc != nil {
		printDocstring(output, t.Doc.Raw)
	}
	for _, a := range t.Ann {
		printAnnotation(output, a)
	}

	lhs := "const " + strutil.ToCamelCase(t.Name) + " = "
	rhs := renderLiteral(t.Value, literalRenderCtx{
		spreadRef:                   refConstDecl,
		scalarRef:                   refConstDecl,
		enumMemberRef:               refEnumMember,
		forceObjectMultiline:        true,
		respectArrayMultilineIntent: true,
		forceCompoundArrayMultiline: true,
	})
	writeRenderedValue(output, lhs, rhs, t.Trailing)
}

func printEnumDecl(output *gen.Generator, t *enumNode) {
	if t.Doc != nil {
		printDocstring(output, t.Doc.Raw)
	}
	for _, a := range t.Ann {
		printAnnotation(output, a)
	}

	name := strutil.ToPascalCase(t.Name)
	if len(t.Members) == 0 {
		lineWithTrailing(output, "enum "+name+" {}", t.Trailing)
		return
	}

	output.Line("enum " + name + " {")
	output.Block(func() {
		for i, m := range t.Members {
			if i > 0 && hasEnumMemberBlankLine(t.Members[i-1], m) {
				blankLine(output)
			}
			printEnumMember(output, m)
		}
	})
	output.Line("}")
}

func printEnumMember(output *gen.Generator, m *enumMemberNode) {
	if m.Comment != nil {
		output.Line(m.Comment.Text)
		return
	}
	if m.Doc != nil {
		printDocstring(output, m.Doc.Raw)
	}
	for _, a := range m.Ann {
		printAnnotation(output, a)
	}
	if m.Spread != nil {
		lineWithTrailing(output, "..."+renderReference(*m.Spread, refEnumDecl), m.Trailing)
		return
	}
	if m.Name == "" {
		return
	}

	line := strutil.ToPascalCase(m.Name)
	if m.Value != nil {
		if m.Value.Str != nil {
			line += ` = "` + strutil.EscapeQuotes(*m.Value.Str) + `"`
		} else if m.Value.Int != nil {
			line += " = " + *m.Value.Int
		}
	}
	lineWithTrailing(output, line, m.Trailing)
}

func printAnnotation(output *gen.Generator, a *annotationNode) {
	name := strutil.ToCamelCase(a.Name)
	if a.Arg == nil {
		output.Line("@" + name)
		return
	}

	renderedArg := renderLiteral(*a.Arg, literalRenderCtx{spreadRef: refConstDecl, scalarRef: refConstDecl, enumMemberRef: refEnumMember})
	if !strings.Contains(renderedArg, "\n") {
		output.Line("@" + name + "(" + renderedArg + ")")
		return
	}

	lines := strings.Split(renderedArg, "\n")
	output.Line("@" + name + "(" + lines[0])
	for i := 1; i < len(lines)-1; i++ {
		if lines[i] == "" {
			blankLine(output)
			continue
		}
		output.Line(lines[i])
	}
	output.Line(lines[len(lines)-1] + ")")
}

func printDocstring(output *gen.Generator, raw string) {
	content := strings.TrimSuffix(strings.TrimPrefix(raw, `"""`), `"""`)
	if !strings.Contains(raw, "\n") {
		output.Line(`""" ` + strings.TrimSpace(content) + ` """`)
		return
	}

	normalized := strutil.NormalizeIndent(content)
	lines := strings.Split(normalized, "\n")
	if len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	output.Line(`"""`)
	for _, line := range lines {
		output.Line(line)
	}
	output.Line(`"""`)
}

func printTypeMember(output *gen.Generator, m *typeMemberNode) {
	switch {
	case m.Comment != nil:
		output.Line(m.Comment.Text)
	case m.Standalone != nil:
		printDocstring(output, m.Standalone.Raw)
	case m.Spread != nil:
		lineWithTrailing(output, "..."+renderReference(*m.Spread, refTypeDecl), m.Trailing)
	case m.Field != nil:
		if m.Trailing != nil && m.Field.Trailing == nil {
			m.Field.Trailing = m.Trailing
		}
		printField(output, m.Field)
	}
}

func printField(output *gen.Generator, f *fieldNode) {
	if f.Doc != nil {
		printDocstring(output, f.Doc.Raw)
	}
	for _, a := range f.Ann {
		printAnnotation(output, a)
	}

	name := strutil.ToCamelCase(f.Name)
	if f.Optional {
		name += "?"
	}
	if f.Type.Obj == nil {
		writeRenderedValue(output, name+" ", renderFieldType(f.Type), f.Trailing)
		return
	}
	if len(f.Type.Obj.Members) == 0 {
		lineWithTrailing(output, name+" {}"+strings.Repeat("[]", f.Type.Dims), f.Trailing)
		return
	}

	output.Line(name + " {")
	output.Block(func() {
		for i, m := range f.Type.Obj.Members {
			if i > 0 && hasTypeMemberBlankLine(f.Type.Obj.Members[i-1], m) {
				blankLine(output)
			}
			printTypeMember(output, m)
		}
	})
	lineWithTrailing(output, "}"+strings.Repeat("[]", f.Type.Dims), f.Trailing)
}

func hasTypeMemberBlankLine(prev, curr *typeMemberNode) bool {
	if prev == nil || curr == nil {
		return false
	}
	return typeMemberStartLine(curr)-typeMemberEndLine(prev) > 1
}

func hasEnumMemberBlankLine(prev, curr *enumMemberNode) bool {
	if prev == nil || curr == nil {
		return false
	}
	return enumMemberStartLine(curr)-enumMemberEndLine(prev) > 1
}

func topNodeStartLine(n node) int {
	switch t := n.(type) {
	case *includeNode:
		return declarationStartLine(t.Doc, t.Ann, t.startLine())
	case *typeNode:
		return declarationStartLine(t.Doc, t.Ann, t.startLine())
	case *constNode:
		return declarationStartLine(t.Doc, t.Ann, t.startLine())
	case *enumNode:
		return declarationStartLine(t.Doc, t.Ann, t.startLine())
	default:
		return n.startLine()
	}
}

func topNodeEndLine(n node) int {
	if n == nil {
		return 0
	}
	return n.endLine()
}

func declarationStartLine(doc *docstringNode, anns []*annotationNode, fallback int) int {
	start := fallback
	if len(anns) > 0 {
		start = min(start, anns[0].startLine())
	}
	if doc != nil {
		start = min(start, doc.startLine())
	}
	return start
}

func typeMemberStartLine(m *typeMemberNode) int {
	if m == nil {
		return 0
	}
	if m.Field != nil {
		return declarationStartLine(m.Field.Doc, m.Field.Ann, m.Field.startLine())
	}
	return m.startLine()
}

func typeMemberEndLine(m *typeMemberNode) int {
	if m == nil {
		return 0
	}
	if m.Field != nil {
		return m.Field.endLine()
	}
	return m.endLine()
}

func enumMemberStartLine(m *enumMemberNode) int {
	if m == nil {
		return 0
	}
	return declarationStartLine(m.Doc, m.Ann, m.startLine())
}

func enumMemberEndLine(m *enumMemberNode) int {
	if m == nil {
		return 0
	}
	return m.endLine()
}
