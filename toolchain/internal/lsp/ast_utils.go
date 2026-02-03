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
// Returns the identifier name, or empty string if not found.
func findIdentifierAtPosition(content string, lspPosition TextDocumentPosition) string {
	lines := strings.Split(content, "\n")
	if lspPosition.Line >= len(lines) {
		return ""
	}

	line := lines[lspPosition.Line]
	if lspPosition.Character >= len(line) {
		return ""
	}

	// Find the start of the identifier
	start := lspPosition.Character
	for start > 0 && isIdentifierChar(line[start-1]) {
		start--
	}

	// Find the end of the identifier
	end := lspPosition.Character
	for end < len(line) && isIdentifierChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}

	return line[start:end]
}

// isIdentifierChar returns true if the character is valid in an identifier.
func isIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// findDocstringPathAtPosition finds if the position is inside a docstring that references an external file.
// Returns the path and true if found, empty string and false otherwise.
func findDocstringPathAtPosition(content string, lspPosition TextDocumentPosition) (string, bool) {
	lines := strings.Split(content, "\n")
	if lspPosition.Line >= len(lines) {
		return "", false
	}

	line := lines[lspPosition.Line]

	// Check if we're inside a docstring that contains a .md path
	// Look for """ markers
	docStart := strings.Index(line, `"""`)
	if docStart == -1 {
		return "", false
	}

	// Find the closing """
	docEnd := strings.Index(line[docStart+3:], `"""`)
	if docEnd == -1 {
		return "", false
	}
	docEnd += docStart + 3

	// Check if cursor is within the docstring
	if lspPosition.Character < docStart || lspPosition.Character > docEnd+3 {
		return "", false
	}

	// Extract content between the quotes
	docContent := strings.TrimSpace(line[docStart+3 : docEnd])

	// Check if it's an external file reference
	if strings.HasSuffix(docContent, ".md") && !strings.ContainsAny(docContent, "\r\n") {
		return docContent, true
	}

	return "", false
}

// collectAllIdentifiersFromSchema collects all identifier occurrences from the parsed schema.
// This walks the AST and collects all identifiers with their positions.
func collectAllIdentifiersFromSchema(schema *ast.Schema, content string) []IdentifierInfo {
	// Pre-allocate a reasonable capacity to avoid frequent re-allocations
	// Most files have < 1000 identifiers
	identifiers := make([]IdentifierInfo, 0, 100)

	for _, child := range schema.Children {
		switch child.Kind() {
		case ast.SchemaChildKindType:
			collectIdentifiersFromType(child.Type, content, &identifiers)
		case ast.SchemaChildKindEnum:
			collectIdentifiersFromEnum(child.Enum, content, &identifiers)
		case ast.SchemaChildKindConst:
			collectIdentifiersFromConst(child.Const, content, &identifiers)
		case ast.SchemaChildKindPattern:
			collectIdentifiersFromPattern(child.Pattern, content, &identifiers)
		case ast.SchemaChildKindRPC:
			collectIdentifiersFromRPC(child.RPC, content, &identifiers)
		}
	}

	return identifiers
}

func collectIdentifiersFromType(t *ast.TypeDecl, content string, identifiers *[]IdentifierInfo) {
	// Type name
	// Find the position of the name
	startPos := t.Pos
	if t.Docstring != nil {
		startPos = t.Docstring.EndPos
	}
	if t.Deprecated != nil {
		startPos = t.Deprecated.EndPos
	}

	namePos, nameEndPos := findIdentifierRange(content, startPos, t.Name)

	*identifiers = append(*identifiers, IdentifierInfo{
		Name:   t.Name,
		Pos:    namePos,
		EndPos: nameEndPos,
	})

	// Fields and spreads
	for _, child := range t.Children {
		if child.Field != nil {
			collectIdentifiersFromField(child.Field, content, identifiers)
		}
		if child.Spread != nil {
			namePos, nameEndPos := findIdentifierRange(content, child.Spread.Pos, child.Spread.TypeName)
			*identifiers = append(*identifiers, IdentifierInfo{
				Name:   child.Spread.TypeName,
				Pos:    namePos,
				EndPos: nameEndPos,
			})
		}
	}
}

func collectIdentifiersFromField(f *ast.Field, content string, identifiers *[]IdentifierInfo) {
	// Field name
	startPos := f.Pos
	if f.Docstring != nil {
		startPos = f.Docstring.EndPos
	}

	namePos, nameEndPos := findIdentifierRange(content, startPos, string(f.Name))

	*identifiers = append(*identifiers, IdentifierInfo{
		Name:   string(f.Name),
		Pos:    namePos,
		EndPos: nameEndPos,
	})

	// Field type references
	collectIdentifiersFromFieldType(&f.Type, content, identifiers)
}

func collectIdentifiersFromFieldType(ft *ast.FieldType, content string, identifiers *[]IdentifierInfo) {
	if ft.Base == nil {
		return
	}

	// Named type reference
	if ft.Base.Named != nil {
		name := *ft.Base.Named
		// Only add if it's not a primitive
		if !ast.IsPrimitiveType(name) {
			// ft.Base.Pos covers the name
			namePos, nameEndPos := findIdentifierRange(content, ft.Base.Pos, name)
			*identifiers = append(*identifiers, IdentifierInfo{
				Name:   name,
				Pos:    namePos,
				EndPos: nameEndPos,
			})
		}
	}

	// Map value type
	if ft.Base.Map != nil && ft.Base.Map.ValueType != nil {
		collectIdentifiersFromFieldType(ft.Base.Map.ValueType, content, identifiers)
	}

	// Inline object
	if ft.Base.Object != nil {
		for _, child := range ft.Base.Object.Children {
			if child.Field != nil {
				collectIdentifiersFromField(child.Field, content, identifiers)
			}
			if child.Spread != nil {
				namePos, nameEndPos := findIdentifierRange(content, child.Spread.Pos, child.Spread.TypeName)
				*identifiers = append(*identifiers, IdentifierInfo{
					Name:   child.Spread.TypeName,
					Pos:    namePos,
					EndPos: nameEndPos,
				})
			}
		}
	}
}

func collectIdentifiersFromEnum(e *ast.EnumDecl, content string, identifiers *[]IdentifierInfo) {
	// Enum name
	startPos := e.Pos
	if e.Docstring != nil {
		startPos = e.Docstring.EndPos
	}
	if e.Deprecated != nil {
		startPos = e.Deprecated.EndPos
	}
	namePos, nameEndPos := findIdentifierRange(content, startPos, e.Name)

	*identifiers = append(*identifiers, IdentifierInfo{
		Name:   e.Name,
		Pos:    namePos,
		EndPos: nameEndPos,
	})

	// Enum members
	for _, member := range e.Members {
		if member.Name != "" {
			namePos, nameEndPos := findIdentifierRange(content, member.Pos, member.Name)
			*identifiers = append(*identifiers, IdentifierInfo{
				Name:   member.Name,
				Pos:    namePos,
				EndPos: nameEndPos,
			})
		}
	}
}

func collectIdentifiersFromConst(c *ast.ConstDecl, content string, identifiers *[]IdentifierInfo) {
	startPos := c.Pos
	if c.Docstring != nil {
		startPos = c.Docstring.EndPos
	}
	if c.Deprecated != nil {
		startPos = c.Deprecated.EndPos
	}
	namePos, nameEndPos := findIdentifierRange(content, startPos, c.Name)

	*identifiers = append(*identifiers, IdentifierInfo{
		Name:   c.Name,
		Pos:    namePos,
		EndPos: nameEndPos,
	})
}

func collectIdentifiersFromPattern(p *ast.PatternDecl, content string, identifiers *[]IdentifierInfo) {
	startPos := p.Pos
	if p.Docstring != nil {
		startPos = p.Docstring.EndPos
	}
	if p.Deprecated != nil {
		startPos = p.Deprecated.EndPos
	}
	namePos, nameEndPos := findIdentifierRange(content, startPos, p.Name)
	*identifiers = append(*identifiers, IdentifierInfo{
		Name:   p.Name,
		Pos:    namePos,
		EndPos: nameEndPos,
	})
}

func collectIdentifiersFromRPC(r *ast.RPCDecl, content string, identifiers *[]IdentifierInfo) {
	// RPC name
	startPos := r.Pos
	if r.Docstring != nil {
		startPos = r.Docstring.EndPos
	}
	if r.Deprecated != nil {
		startPos = r.Deprecated.EndPos
	}
	namePos, nameEndPos := findIdentifierRange(content, startPos, r.Name)

	*identifiers = append(*identifiers, IdentifierInfo{
		Name:   r.Name,
		Pos:    namePos,
		EndPos: nameEndPos,
	})

	// Procs and Streams
	for _, child := range r.Children {
		if child.Proc != nil {
			collectIdentifiersFromProc(child.Proc, content, identifiers)
		}
		if child.Stream != nil {
			collectIdentifiersFromStream(child.Stream, content, identifiers)
		}
	}
}

func collectIdentifiersFromProc(p *ast.ProcDecl, content string, identifiers *[]IdentifierInfo) {
	// Proc name
	startPos := p.Pos
	if p.Docstring != nil {
		startPos = p.Docstring.EndPos
	}
	if p.Deprecated != nil {
		startPos = p.Deprecated.EndPos
	}
	namePos, nameEndPos := findIdentifierRange(content, startPos, p.Name)

	*identifiers = append(*identifiers, IdentifierInfo{
		Name:   p.Name,
		Pos:    namePos,
		EndPos: nameEndPos,
	})

	// Input/Output
	for _, child := range p.Children {
		if child.Input != nil {
			collectIdentifiersFromInputOutput(child.Input.Children, content, identifiers)
		}
		if child.Output != nil {
			collectIdentifiersFromInputOutput(child.Output.Children, content, identifiers)
		}
	}
}

func collectIdentifiersFromStream(s *ast.StreamDecl, content string, identifiers *[]IdentifierInfo) {
	// Stream name
	startPos := s.Pos
	if s.Docstring != nil {
		startPos = s.Docstring.EndPos
	}
	if s.Deprecated != nil {
		startPos = s.Deprecated.EndPos
	}
	namePos, nameEndPos := findIdentifierRange(content, startPos, s.Name)

	*identifiers = append(*identifiers, IdentifierInfo{
		Name:   s.Name,
		Pos:    namePos,
		EndPos: nameEndPos,
	})

	// Input/Output
	for _, child := range s.Children {
		if child.Input != nil {
			collectIdentifiersFromInputOutput(child.Input.Children, content, identifiers)
		}
		if child.Output != nil {
			collectIdentifiersFromInputOutput(child.Output.Children, content, identifiers)
		}
	}
}

func collectIdentifiersFromInputOutput(children []*ast.InputOutputChild, content string, identifiers *[]IdentifierInfo) {
	for _, child := range children {
		if child.Field != nil {
			collectIdentifiersFromField(child.Field, content, identifiers)
		}
		if child.Spread != nil {
			namePos, nameEndPos := findIdentifierRange(content, child.Spread.Pos, child.Spread.TypeName)
			*identifiers = append(*identifiers, IdentifierInfo{
				Name:   child.Spread.TypeName,
				Pos:    namePos,
				EndPos: nameEndPos,
			})
		}
	}
}

// findReferencesInSchema finds all occurrences of a symbol name in the schema.
func findReferencesInSchema(schema *ast.Schema, symbolName, content string) []IdentifierInfo {
	allIdentifiers := collectAllIdentifiersFromSchema(schema, content)

	var references []IdentifierInfo
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

	// Search for the name starting from startOffset
	idx := strings.Index(content[startOffset:], name)
	if idx == -1 {
		return startSearchPos, startSearchPos
	}

	matchOffset := startOffset + idx

	// Calculate new position
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
// Returns the location of the definition, or nil if not found.
func resolveSymbolDefinition(program *analysis.Program, symbolName string) *Location {
	// Check types
	if t, ok := program.Types[symbolName]; ok {
		return &Location{
			URI: PathToUri(t.File),
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(t.Pos),
				End:   convertASTPositionToLSPPosition(t.EndPos),
			},
		}
	}

	// Check enums
	if e, ok := program.Enums[symbolName]; ok {
		return &Location{
			URI: PathToUri(e.File),
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(e.Pos),
				End:   convertASTPositionToLSPPosition(e.EndPos),
			},
		}
	}

	// Check constants
	if c, ok := program.Consts[symbolName]; ok {
		return &Location{
			URI: PathToUri(c.File),
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(c.Pos),
				End:   convertASTPositionToLSPPosition(c.EndPos),
			},
		}
	}

	// Check patterns
	if p, ok := program.Patterns[symbolName]; ok {
		return &Location{
			URI: PathToUri(p.File),
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(p.Pos),
				End:   convertASTPositionToLSPPosition(p.EndPos),
			},
		}
	}

	// Check RPCs
	if r, ok := program.RPCs[symbolName]; ok {
		return &Location{
			URI: PathToUri(r.File),
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(r.Pos),
				End:   convertASTPositionToLSPPosition(r.EndPos),
			},
		}
	}

	// Check procs and streams within RPCs
	for _, rpc := range program.RPCs {
		if proc, ok := rpc.Procs[symbolName]; ok {
			return &Location{
				URI: PathToUri(proc.File),
				Range: TextDocumentRange{
					Start: convertASTPositionToLSPPosition(proc.Pos),
					End:   convertASTPositionToLSPPosition(proc.EndPos),
				},
			}
		}
		if stream, ok := rpc.Streams[symbolName]; ok {
			return &Location{
				URI: PathToUri(stream.File),
				Range: TextDocumentRange{
					Start: convertASTPositionToLSPPosition(stream.Pos),
					End:   convertASTPositionToLSPPosition(stream.EndPos),
				},
			}
		}
	}

	return nil
}
