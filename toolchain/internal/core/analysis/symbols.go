package analysis

import (
	"fmt"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// symbolTable manages the collection and lookup of symbols during analysis.
// It tracks all declarations and their origins for building the final Program.
type symbolTable struct {
	// All collected symbols
	types  map[string]*TypeSymbol
	enums  map[string]*EnumSymbol
	consts map[string]*ConstSymbol

	// Standalone docstrings at the schema level
	standaloneDocs []*DocSymbol

	// Track where each symbol was first declared (for duplicate detection)
	typeOrigins  map[string]ast.Position
	enumOrigins  map[string]ast.Position
	constOrigins map[string]ast.Position
}

// newSymbolTable creates a new empty symbol table.
func newSymbolTable() *symbolTable {
	return &symbolTable{
		types:          make(map[string]*TypeSymbol),
		enums:          make(map[string]*EnumSymbol),
		consts:         make(map[string]*ConstSymbol),
		standaloneDocs: []*DocSymbol{},
		typeOrigins:    make(map[string]ast.Position),
		enumOrigins:    make(map[string]ast.Position),
		constOrigins:   make(map[string]ast.Position),
	}
}

// registerType attempts to register a type symbol.
// Returns a Diagnostic if the type name is already declared.
func (st *symbolTable) registerType(sym *TypeSymbol) *Diagnostic {
	if existing, ok := st.typeOrigins[sym.Name]; ok {
		diag := newDiagnostic(
			sym.File,
			sym.Pos,
			sym.EndPos,
			CodeDuplicateType,
			formatDuplicateError("type", sym.Name, existing),
		)
		return &diag
	}
	st.types[sym.Name] = sym
	st.typeOrigins[sym.Name] = sym.Pos
	return nil
}

// registerEnum attempts to register an enum symbol.
// Returns a Diagnostic if the enum name is already declared.
func (st *symbolTable) registerEnum(sym *EnumSymbol) *Diagnostic {
	if existing, ok := st.enumOrigins[sym.Name]; ok {
		diag := newDiagnostic(
			sym.File,
			sym.Pos,
			sym.EndPos,
			CodeDuplicateEnum,
			formatDuplicateError("enum", sym.Name, existing),
		)
		return &diag
	}
	st.enums[sym.Name] = sym
	st.enumOrigins[sym.Name] = sym.Pos
	return nil
}

// registerConst attempts to register a const symbol.
// Returns a Diagnostic if the const name is already declared.
func (st *symbolTable) registerConst(sym *ConstSymbol) *Diagnostic {
	if existing, ok := st.constOrigins[sym.Name]; ok {
		diag := newDiagnostic(
			sym.File,
			sym.Pos,
			sym.EndPos,
			CodeDuplicateConst,
			formatDuplicateError("constant", sym.Name, existing),
		)
		return &diag
	}
	st.consts[sym.Name] = sym
	st.constOrigins[sym.Name] = sym.Pos
	return nil
}

// addStandaloneDoc adds a standalone docstring to the symbol table.
func (st *symbolTable) addStandaloneDoc(doc *DocSymbol) {
	st.standaloneDocs = append(st.standaloneDocs, doc)
}

// lookupType returns the type symbol with the given name, or nil if not found.
func (st *symbolTable) lookupType(name string) *TypeSymbol {
	return st.types[name]
}

// lookupEnum returns the enum symbol with the given name, or nil if not found.
func (st *symbolTable) lookupEnum(name string) *EnumSymbol {
	return st.enums[name]
}

// allTypeNames returns a slice of all registered type names.
func (st *symbolTable) allTypeNames() []string {
	names := make([]string, 0, len(st.types))
	for name := range st.types {
		names = append(names, name)
	}
	return names
}

// allFieldTypeNames returns all valid type names for field types.
// This includes primitive types, custom types, and enums.
func (st *symbolTable) allFieldTypeNames() []string {
	names := make([]string, 0, len(st.types)+len(st.enums)+5)

	// Add primitive types
	names = append(names, "string", "int", "float", "bool", "datetime")

	// Add custom types
	for name := range st.types {
		names = append(names, name)
	}

	// Add enums
	for name := range st.enums {
		names = append(names, name)
	}

	return names
}

// allConstNames returns a slice of all registered const names.
func (st *symbolTable) allConstNames() []string {
	names := make([]string, 0, len(st.consts))
	for name := range st.consts {
		names = append(names, name)
	}
	return names
}

// lookupConst returns the const symbol with the given name, or nil if not found.
func (st *symbolTable) lookupConst(name string) *ConstSymbol {
	return st.consts[name]
}

// buildProgram creates a Program from the collected symbols.
func (st *symbolTable) buildProgram(entryPoint string, files map[string]*File) *Program {
	return &Program{
		EntryPoint:     entryPoint,
		Files:          files,
		Types:          st.types,
		Enums:          st.enums,
		Consts:         st.consts,
		StandaloneDocs: st.standaloneDocs,
	}
}

// formatDuplicateError creates a standardized duplicate declaration error message.
func formatDuplicateError(kind, name string, original ast.Position) string {
	return fmt.Sprintf("%s %q is already declared at %s:%d:%d",
		kind, name, original.Filename, original.Line, original.Column)
}
