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
func collectAllIdentifiersFromSchema(schema *ast.Schema) []IdentifierInfo {
	var identifiers []IdentifierInfo

	for _, child := range schema.Children {
		switch child.Kind() {
		case ast.SchemaChildKindType:
			identifiers = append(identifiers, collectIdentifiersFromType(child.Type)...)
		case ast.SchemaChildKindEnum:
			identifiers = append(identifiers, collectIdentifiersFromEnum(child.Enum)...)
		case ast.SchemaChildKindConst:
			identifiers = append(identifiers, collectIdentifiersFromConst(child.Const)...)
		case ast.SchemaChildKindPattern:
			identifiers = append(identifiers, collectIdentifiersFromPattern(child.Pattern)...)
		case ast.SchemaChildKindRPC:
			identifiers = append(identifiers, collectIdentifiersFromRPC(child.RPC)...)
		}
	}

	return identifiers
}

func collectIdentifiersFromType(t *ast.TypeDecl) []IdentifierInfo {
	var identifiers []IdentifierInfo

	// Type name
	identifiers = append(identifiers, IdentifierInfo{
		Name:   t.Name,
		Pos:    t.Pos,
		EndPos: t.EndPos,
	})

	// Fields and spreads
	for _, child := range t.Children {
		if child.Field != nil {
			identifiers = append(identifiers, collectIdentifiersFromField(child.Field)...)
		}
		if child.Spread != nil {
			identifiers = append(identifiers, IdentifierInfo{
				Name:   child.Spread.TypeName,
				Pos:    child.Spread.Pos,
				EndPos: child.Spread.EndPos,
			})
		}
	}

	return identifiers
}

func collectIdentifiersFromField(f *ast.Field) []IdentifierInfo {
	var identifiers []IdentifierInfo

	// Field name
	identifiers = append(identifiers, IdentifierInfo{
		Name:   f.Name,
		Pos:    f.Pos,
		EndPos: f.EndPos,
	})

	// Field type references
	identifiers = append(identifiers, collectIdentifiersFromFieldType(&f.Type)...)

	return identifiers
}

func collectIdentifiersFromFieldType(ft *ast.FieldType) []IdentifierInfo {
	var identifiers []IdentifierInfo

	if ft.Base == nil {
		return identifiers
	}

	// Named type reference
	if ft.Base.Named != nil {
		name := *ft.Base.Named
		// Only add if it's not a primitive
		if !ast.IsPrimitiveType(name) {
			identifiers = append(identifiers, IdentifierInfo{
				Name:   name,
				Pos:    ft.Base.Pos,
				EndPos: ft.Base.EndPos,
			})
		}
	}

	// Map value type
	if ft.Base.Map != nil && ft.Base.Map.ValueType != nil {
		identifiers = append(identifiers, collectIdentifiersFromFieldType(ft.Base.Map.ValueType)...)
	}

	// Inline object
	if ft.Base.Object != nil {
		for _, child := range ft.Base.Object.Children {
			if child.Field != nil {
				identifiers = append(identifiers, collectIdentifiersFromField(child.Field)...)
			}
			if child.Spread != nil {
				identifiers = append(identifiers, IdentifierInfo{
					Name:   child.Spread.TypeName,
					Pos:    child.Spread.Pos,
					EndPos: child.Spread.EndPos,
				})
			}
		}
	}

	return identifiers
}

func collectIdentifiersFromEnum(e *ast.EnumDecl) []IdentifierInfo {
	var identifiers []IdentifierInfo

	// Enum name
	identifiers = append(identifiers, IdentifierInfo{
		Name:   e.Name,
		Pos:    e.Pos,
		EndPos: e.EndPos,
	})

	// Enum members
	for _, member := range e.Members {
		if member.Name != "" {
			identifiers = append(identifiers, IdentifierInfo{
				Name:   member.Name,
				Pos:    member.Pos,
				EndPos: member.EndPos,
			})
		}
	}

	return identifiers
}

func collectIdentifiersFromConst(c *ast.ConstDecl) []IdentifierInfo {
	return []IdentifierInfo{
		{
			Name:   c.Name,
			Pos:    c.Pos,
			EndPos: c.EndPos,
		},
	}
}

func collectIdentifiersFromPattern(p *ast.PatternDecl) []IdentifierInfo {
	return []IdentifierInfo{
		{
			Name:   p.Name,
			Pos:    p.Pos,
			EndPos: p.EndPos,
		},
	}
}

func collectIdentifiersFromRPC(r *ast.RPCDecl) []IdentifierInfo {
	var identifiers []IdentifierInfo

	// RPC name
	identifiers = append(identifiers, IdentifierInfo{
		Name:   r.Name,
		Pos:    r.Pos,
		EndPos: r.EndPos,
	})

	// Procs and Streams
	for _, child := range r.Children {
		if child.Proc != nil {
			identifiers = append(identifiers, collectIdentifiersFromProc(child.Proc)...)
		}
		if child.Stream != nil {
			identifiers = append(identifiers, collectIdentifiersFromStream(child.Stream)...)
		}
	}

	return identifiers
}

func collectIdentifiersFromProc(p *ast.ProcDecl) []IdentifierInfo {
	var identifiers []IdentifierInfo

	// Proc name
	identifiers = append(identifiers, IdentifierInfo{
		Name:   p.Name,
		Pos:    p.Pos,
		EndPos: p.EndPos,
	})

	// Input/Output
	for _, child := range p.Children {
		if child.Input != nil {
			identifiers = append(identifiers, collectIdentifiersFromInputOutput(child.Input.Children)...)
		}
		if child.Output != nil {
			identifiers = append(identifiers, collectIdentifiersFromInputOutput(child.Output.Children)...)
		}
	}

	return identifiers
}

func collectIdentifiersFromStream(s *ast.StreamDecl) []IdentifierInfo {
	var identifiers []IdentifierInfo

	// Stream name
	identifiers = append(identifiers, IdentifierInfo{
		Name:   s.Name,
		Pos:    s.Pos,
		EndPos: s.EndPos,
	})

	// Input/Output
	for _, child := range s.Children {
		if child.Input != nil {
			identifiers = append(identifiers, collectIdentifiersFromInputOutput(child.Input.Children)...)
		}
		if child.Output != nil {
			identifiers = append(identifiers, collectIdentifiersFromInputOutput(child.Output.Children)...)
		}
	}

	return identifiers
}

func collectIdentifiersFromInputOutput(children []*ast.InputOutputChild) []IdentifierInfo {
	var identifiers []IdentifierInfo

	for _, child := range children {
		if child.Field != nil {
			identifiers = append(identifiers, collectIdentifiersFromField(child.Field)...)
		}
		if child.Spread != nil {
			identifiers = append(identifiers, IdentifierInfo{
				Name:   child.Spread.TypeName,
				Pos:    child.Spread.Pos,
				EndPos: child.Spread.EndPos,
			})
		}
	}

	return identifiers
}

// findReferencesInSchema finds all occurrences of a symbol name in the schema.
func findReferencesInSchema(schema *ast.Schema, symbolName string) []IdentifierInfo {
	allIdentifiers := collectAllIdentifiersFromSchema(schema)

	var references []IdentifierInfo
	for _, id := range allIdentifiers {
		if id.Name == symbolName {
			references = append(references, id)
		}
	}

	return references
}

// resolveSymbolDefinition finds the definition of a symbol in the program.
// Returns the location of the definition, or nil if not found.
func resolveSymbolDefinition(program *analysis.Program, symbolName string) *Location {
	// Check types
	if t, ok := program.Types[symbolName]; ok {
		return &Location{
			URI: pathToURI(t.File),
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(t.Pos),
				End:   convertASTPositionToLSPPosition(t.EndPos),
			},
		}
	}

	// Check enums
	if e, ok := program.Enums[symbolName]; ok {
		return &Location{
			URI: pathToURI(e.File),
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(e.Pos),
				End:   convertASTPositionToLSPPosition(e.EndPos),
			},
		}
	}

	// Check constants
	if c, ok := program.Consts[symbolName]; ok {
		return &Location{
			URI: pathToURI(c.File),
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(c.Pos),
				End:   convertASTPositionToLSPPosition(c.EndPos),
			},
		}
	}

	// Check patterns
	if p, ok := program.Patterns[symbolName]; ok {
		return &Location{
			URI: pathToURI(p.File),
			Range: TextDocumentRange{
				Start: convertASTPositionToLSPPosition(p.Pos),
				End:   convertASTPositionToLSPPosition(p.EndPos),
			},
		}
	}

	// Check RPCs
	if r, ok := program.RPCs[symbolName]; ok {
		return &Location{
			URI: pathToURI(r.File),
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
				URI: pathToURI(proc.File),
				Range: TextDocumentRange{
					Start: convertASTPositionToLSPPosition(proc.Pos),
					End:   convertASTPositionToLSPPosition(proc.EndPos),
				},
			}
		}
		if stream, ok := rpc.Streams[symbolName]; ok {
			return &Location{
				URI: pathToURI(stream.File),
				Range: TextDocumentRange{
					Start: convertASTPositionToLSPPosition(stream.Pos),
					End:   convertASTPositionToLSPPosition(stream.EndPos),
				},
			}
		}
	}

	return nil
}
