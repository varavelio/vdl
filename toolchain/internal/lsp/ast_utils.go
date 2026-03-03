package lsp

import (
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/core/analysis"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// IdentifierInfo represents information about an identifier found in the AST.
type IdentifierInfo struct {
	Name   string
	Pos    ast.Position
	EndPos ast.Position
}

// findIdentifierAtPosition finds the identifier at the given LSP position in the content.
func findIdentifierAtPosition(content string, lspPosition TextDocumentPosition) string {
	lines := strings.Split(content, "\n")
	if lspPosition.Line >= len(lines) {
		return ""
	}

	line := lines[lspPosition.Line]
	if lspPosition.Character >= len(line) {
		return ""
	}

	start := lspPosition.Character
	for start > 0 && isIdentifierChar(line[start-1]) {
		start--
	}

	end := lspPosition.Character
	for end < len(line) && isIdentifierChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}

	return line[start:end]
}

func isIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// findDocstringPathAtPosition finds if the position is inside a single-line docstring
// that references an external .md file.
func findDocstringPathAtPosition(content string, lspPosition TextDocumentPosition) (string, bool) {
	lines := strings.Split(content, "\n")
	if lspPosition.Line >= len(lines) {
		return "", false
	}

	line := lines[lspPosition.Line]
	docStart := strings.Index(line, `"""`)
	if docStart == -1 {
		return "", false
	}

	docEnd := strings.Index(line[docStart+3:], `"""`)
	if docEnd == -1 {
		return "", false
	}
	docEnd += docStart + 3

	if lspPosition.Character < docStart || lspPosition.Character > docEnd+3 {
		return "", false
	}

	docContent := strings.TrimSpace(line[docStart+3 : docEnd])
	if strings.HasSuffix(docContent, ".md") && !strings.ContainsAny(docContent, "\r\n") {
		return docContent, true
	}

	return "", false
}

func collectAllIdentifiersFromSchema(schema *ast.Schema, content string) []IdentifierInfo {
	ids := make([]IdentifierInfo, 0, 128)

	for _, decl := range schema.Declarations {
		switch decl.Kind() {
		case ast.DeclKindType:
			collectIdentifiersFromType(decl.Type, content, &ids)
		case ast.DeclKindEnum:
			collectIdentifiersFromEnum(decl.Enum, content, &ids)
		case ast.DeclKindConst:
			collectIdentifiersFromConst(decl.Const, content, &ids)
		}
	}

	return ids
}

func collectIdentifiersFromType(t *ast.TypeDecl, content string, ids *[]IdentifierInfo) {
	if t == nil {
		return
	}

	namePos, nameEndPos := findIdentifierRange(content, t.Pos, t.Name)
	*ids = append(*ids, IdentifierInfo{Name: t.Name, Pos: namePos, EndPos: nameEndPos})

	collectIdentifiersFromAnnotations(t.Annotations, content, ids)
	if !t.IsObject() {
		typeRef := t.Type()
		collectIdentifiersFromFieldType(&typeRef, content, ids)
	}

	for _, m := range t.Members() {
		collectIdentifiersFromTypeMember(m, content, ids)
	}
}

func collectIdentifiersFromTypeMember(m *ast.TypeMember, content string, ids *[]IdentifierInfo) {
	if m == nil {
		return
	}
	if m.Field != nil {
		collectIdentifiersFromField(m.Field, content, ids)
	}
	if m.Spread != nil {
		collectIdentifiersFromSpread(m.Spread, content, ids)
	}
}

func collectIdentifiersFromField(f *ast.Field, content string, ids *[]IdentifierInfo) {
	if f == nil {
		return
	}

	namePos, nameEndPos := findIdentifierRange(content, f.Pos, f.Name)
	*ids = append(*ids, IdentifierInfo{Name: f.Name, Pos: namePos, EndPos: nameEndPos})

	collectIdentifiersFromAnnotations(f.Annotations, content, ids)
	collectIdentifiersFromFieldType(&f.Type, content, ids)
}

func collectIdentifiersFromFieldType(ft *ast.FieldType, content string, ids *[]IdentifierInfo) {
	if ft == nil || ft.Base == nil {
		return
	}

	if ft.Base.Named != nil {
		name := *ft.Base.Named
		if !ast.IsPrimitiveType(name) {
			namePos, nameEndPos := findIdentifierRange(content, ft.Base.Pos, name)
			*ids = append(*ids, IdentifierInfo{Name: name, Pos: namePos, EndPos: nameEndPos})
		}
	}

	if ft.Base.Map != nil {
		collectIdentifiersFromFieldType(ft.Base.Map.ValueType, content, ids)
	}

	if ft.Base.Object != nil {
		for _, m := range ft.Base.Object.Members {
			collectIdentifiersFromTypeMember(m, content, ids)
		}
	}
}

func collectIdentifiersFromEnum(e *ast.EnumDecl, content string, ids *[]IdentifierInfo) {
	if e == nil {
		return
	}

	namePos, nameEndPos := findIdentifierRange(content, e.Pos, e.Name)
	*ids = append(*ids, IdentifierInfo{Name: e.Name, Pos: namePos, EndPos: nameEndPos})

	collectIdentifiersFromAnnotations(e.Annotations, content, ids)

	for _, m := range e.Members {
		if m == nil {
			continue
		}
		if m.Spread != nil {
			collectIdentifiersFromSpread(m.Spread, content, ids)
			continue
		}
		if m.Name != "" {
			memberPos, memberEndPos := findIdentifierRange(content, m.Pos, m.Name)
			*ids = append(*ids, IdentifierInfo{Name: m.Name, Pos: memberPos, EndPos: memberEndPos})
		}
		collectIdentifiersFromAnnotations(m.Annotations, content, ids)
	}
}

func collectIdentifiersFromConst(c *ast.ConstDecl, content string, ids *[]IdentifierInfo) {
	if c == nil {
		return
	}

	namePos, nameEndPos := findIdentifierRange(content, c.Pos, c.Name)
	*ids = append(*ids, IdentifierInfo{Name: c.Name, Pos: namePos, EndPos: nameEndPos})

	collectIdentifiersFromAnnotations(c.Annotations, content, ids)
	collectIdentifiersFromDataLiteral(c.Value, content, ids)
}

func collectIdentifiersFromAnnotations(annotations []*ast.Annotation, content string, ids *[]IdentifierInfo) {
	for _, ann := range annotations {
		if ann == nil || ann.Argument == nil {
			continue
		}
		collectIdentifiersFromDataLiteral(ann.Argument, content, ids)
	}
}

func collectIdentifiersFromDataLiteral(lit *ast.DataLiteral, content string, ids *[]IdentifierInfo) {
	if lit == nil {
		return
	}

	if lit.Scalar != nil && lit.Scalar.Ref != nil {
		collectIdentifiersFromReference(lit.Scalar.Ref, content, ids)
	}

	if lit.Object != nil {
		for _, e := range lit.Object.Entries {
			if e == nil {
				continue
			}
			if e.Spread != nil {
				collectIdentifiersFromSpread(e.Spread, content, ids)
			}
			collectIdentifiersFromDataLiteral(e.Value, content, ids)
		}
	}

	if lit.Array != nil {
		for _, el := range lit.Array.Elements {
			collectIdentifiersFromDataLiteral(el, content, ids)
		}
	}
}

func collectIdentifiersFromSpread(s *ast.Spread, content string, ids *[]IdentifierInfo) {
	if s == nil || s.Ref == nil {
		return
	}
	collectIdentifiersFromReference(s.Ref, content, ids)
}

func collectIdentifiersFromReference(ref *ast.Reference, content string, ids *[]IdentifierInfo) {
	if ref == nil {
		return
	}

	namePos, nameEndPos := findIdentifierRange(content, ref.Pos, ref.Name)
	*ids = append(*ids, IdentifierInfo{Name: ref.Name, Pos: namePos, EndPos: nameEndPos})

	if ref.Member != nil {
		memberPos, memberEndPos := findIdentifierRange(content, nameEndPos, *ref.Member)
		*ids = append(*ids, IdentifierInfo{Name: *ref.Member, Pos: memberPos, EndPos: memberEndPos})
	}
}

// findReferencesInSchema finds all occurrences of a symbol name in the schema.
func findReferencesInSchema(schema *ast.Schema, symbolName, content string) []IdentifierInfo {
	allIdentifiers := collectAllIdentifiersFromSchema(schema, content)

	references := make([]IdentifierInfo, 0, len(allIdentifiers))
	for _, id := range allIdentifiers {
		if id.Name == symbolName {
			references = append(references, id)
		}
	}

	return references
}

// findIdentifierRange finds the precise start and end position of an identifier in content
// starting search from startSearchPos.
func findIdentifierRange(content string, startSearchPos ast.Position, name string) (ast.Position, ast.Position) {
	startOffset := startSearchPos.Offset
	if startOffset >= len(content) {
		return startSearchPos, startSearchPos
	}

	idx := strings.Index(content[startOffset:], name)
	if idx == -1 {
		return startSearchPos, startSearchPos
	}

	matchOffset := startOffset + idx

	newPos := startSearchPos
	segment := content[startOffset:matchOffset]
	newlines := strings.Count(segment, "\n")
	newPos.Line += newlines
	if newlines > 0 {
		lastNewlineIdx := strings.LastIndex(segment, "\n")
		newPos.Column = len(segment) - lastNewlineIdx
	} else {
		newPos.Column += len(segment)
	}
	newPos.Offset = matchOffset

	endPos := newPos
	endPos.Column += len(name)
	endPos.Offset += len(name)

	return newPos, endPos
}

// resolveSymbolDefinition finds the definition of a symbol in the program.
func resolveSymbolDefinition(program *analysis.Program, symbolName string) *Location {
	if program == nil {
		return nil
	}

	if t, ok := program.Types[symbolName]; ok {
		return symbolToLocation(t.File, t.Pos, t.EndPos)
	}
	if e, ok := program.Enums[symbolName]; ok {
		return symbolToLocation(e.File, e.Pos, e.EndPos)
	}
	if c, ok := program.Consts[symbolName]; ok {
		return symbolToLocation(c.File, c.Pos, c.EndPos)
	}

	// Allow go-to-definition on enum member names.
	for _, e := range program.Enums {
		for _, m := range e.Members {
			if m.Name == symbolName {
				return symbolToLocation(m.File, m.Pos, m.EndPos)
			}
		}
	}

	return nil
}

func symbolToLocation(file string, pos, endPos ast.Position) *Location {
	return &Location{
		URI: PathToUri(file),
		Range: TextDocumentRange{
			Start: convertASTPositionToLSPPosition(pos),
			End:   convertASTPositionToLSPPosition(endPos),
		},
	}
}
