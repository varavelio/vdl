package formatter

import (
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

func printDocument(w *fmtWriter, d *docNode) {
	for i, item := range d.Items {
		if i > 0 && hasTopLevelBlankLine(d.Items[i-1], item) {
			w.blank()
		}
		printTopNode(w, item)
	}
}

func hasTopLevelBlankLine(prev, curr node) bool {
	if prev == nil || curr == nil {
		return false
	}
	return topNodeStartLine(curr)-topNodeEndLine(prev) > 1
}

func printTopNode(w *fmtWriter, n node) {
	switch t := n.(type) {
	case *commentNode:
		w.line(t.Text)
	case *docstringNode:
		printDocstring(w, t.Raw)
	case *includeNode:
		if t.Doc != nil {
			printDocstring(w, t.Doc.Raw)
		}
		for _, a := range t.Ann {
			printAnnotation(w, a)
		}
		w.lineWithTrailing(`include "`+strutil.EscapeQuotes(t.Path)+`"`, t.Trailing)
	case *typeNode:
		printTypeDecl(w, t)
	case *constNode:
		printConstDecl(w, t)
	case *enumNode:
		printEnumDecl(w, t)
	}
}

func printTypeDecl(w *fmtWriter, t *typeNode) {
	if t.Doc != nil {
		printDocstring(w, t.Doc.Raw)
	}
	for _, a := range t.Ann {
		printAnnotation(w, a)
	}
	name := strutil.ToPascalCase(t.Name)
	if t.Type.Obj == nil {
		w.lineWithTrailing("type "+name+" "+renderFieldType(t.Type), t.Trailing)
		return
	}
	if len(t.Type.Obj.Members) == 0 {
		w.lineWithTrailing("type "+name+" {}"+strings.Repeat("[]", t.Type.Dims), t.Trailing)
		return
	}
	w.line("type " + name + " {")
	w.indent++
	for i, m := range t.Type.Obj.Members {
		if i > 0 && hasTypeMemberBlankLine(t.Type.Obj.Members[i-1], m) {
			w.blank()
		}
		printTypeMember(w, m)
	}
	w.indent--
	w.lineWithTrailing("}"+strings.Repeat("[]", t.Type.Dims), t.Trailing)
}

func printConstDecl(w *fmtWriter, t *constNode) {
	if t.Doc != nil {
		printDocstring(w, t.Doc.Raw)
	}
	for _, a := range t.Ann {
		printAnnotation(w, a)
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
	printMultilineStatement(w, lhs, rhs, t.Trailing)
}

func printEnumDecl(w *fmtWriter, t *enumNode) {
	if t.Doc != nil {
		printDocstring(w, t.Doc.Raw)
	}
	for _, a := range t.Ann {
		printAnnotation(w, a)
	}
	name := strutil.ToPascalCase(t.Name)
	if len(t.Members) == 0 {
		w.lineWithTrailing("enum "+name+" {}", t.Trailing)
		return
	}
	w.line("enum " + name + " {")
	w.indent++
	for i, m := range t.Members {
		if i > 0 && hasEnumMemberBlankLine(t.Members[i-1], m) {
			w.blank()
		}
		printEnumMember(w, m)
	}
	w.indent--
	w.line("}")
}

func printEnumMember(w *fmtWriter, m *enumMemberNode) {
	if m.Comment != nil {
		w.line(m.Comment.Text)
		return
	}
	if m.Doc != nil {
		printDocstring(w, m.Doc.Raw)
	}
	for _, a := range m.Ann {
		printAnnotation(w, a)
	}
	if m.Spread != nil {
		w.lineWithTrailing("..."+renderReference(*m.Spread, refEnumDecl), m.Trailing)
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
	w.lineWithTrailing(line, m.Trailing)
}

func printAnnotation(w *fmtWriter, a *annotationNode) {
	name := strutil.ToCamelCase(a.Name)
	if a.Arg == nil {
		w.line("@" + name)
		return
	}
	renderedArg := renderLiteral(*a.Arg, literalRenderCtx{spreadRef: refConstDecl, scalarRef: refConstDecl, enumMemberRef: refEnumMember})
	if !strings.Contains(renderedArg, "\n") {
		w.line("@" + name + "(" + renderedArg + ")")
		return
	}
	lines := strings.Split(renderedArg, "\n")
	w.line("@" + name + "(" + lines[0])
	for i := 1; i < len(lines)-1; i++ {
		if lines[i] == "" {
			w.blank()
			continue
		}
		w.line(lines[i])
	}
	w.line(lines[len(lines)-1] + ")")
}

func printDocstring(w *fmtWriter, raw string) {
	content := strings.TrimSuffix(strings.TrimPrefix(raw, `"""`), `"""`)
	if !strings.Contains(raw, "\n") {
		w.line(`""" ` + strings.TrimSpace(content) + ` """`)
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
	w.line(`"""`)
	for _, l := range lines {
		w.line(l)
	}
	w.line(`"""`)
}

func printTypeMember(w *fmtWriter, m *typeMemberNode) {
	switch {
	case m.Comment != nil:
		w.line(m.Comment.Text)
	case m.Standalone != nil:
		printDocstring(w, m.Standalone.Raw)
	case m.Spread != nil:
		w.lineWithTrailing("..."+renderReference(*m.Spread, refTypeDecl), m.Trailing)
	case m.Field != nil:
		if m.Trailing != nil && m.Field.Trailing == nil {
			m.Field.Trailing = m.Trailing
		}
		printField(w, m.Field)
	}
}

func printField(w *fmtWriter, f *fieldNode) {
	if f.Doc != nil {
		printDocstring(w, f.Doc.Raw)
	}
	for _, a := range f.Ann {
		printAnnotation(w, a)
	}
	name := strutil.ToCamelCase(f.Name)
	if f.Optional {
		name += "?"
	}
	if f.Type.Obj == nil {
		w.lineWithTrailing(name+" "+renderFieldType(f.Type), f.Trailing)
		return
	}
	if len(f.Type.Obj.Members) == 0 {
		w.lineWithTrailing(name+" {}"+strings.Repeat("[]", f.Type.Dims), f.Trailing)
		return
	}
	w.line(name + " {")
	w.indent++
	for i, m := range f.Type.Obj.Members {
		if i > 0 && hasTypeMemberBlankLine(f.Type.Obj.Members[i-1], m) {
			w.blank()
		}
		printTypeMember(w, m)
	}
	w.indent--
	w.lineWithTrailing("}"+strings.Repeat("[]", f.Type.Dims), f.Trailing)
}

func printMultilineStatement(w *fmtWriter, lhs, rhs string, trailing *commentNode) {
	writeRenderedValue(w, lhs, rhs, trailing)
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
