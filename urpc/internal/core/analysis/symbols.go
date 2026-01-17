package analysis

import (
	"fmt"

	"github.com/uforg/uforpc/urpc/internal/core/ast"
)

// symbolTable manages the collection and lookup of symbols during analysis.
// It tracks all declarations and their origins for building the final Program.
type symbolTable struct {
	// All collected symbols
	types    map[string]*TypeSymbol
	enums    map[string]*EnumSymbol
	consts   map[string]*ConstSymbol
	patterns map[string]*PatternSymbol
	rpcs     map[string]*RPCSymbol

	// Standalone docstrings at the schema level
	standaloneDocs []*DocSymbol

	// Track where each symbol was first declared (for duplicate detection)
	typeOrigins    map[string]ast.Position
	enumOrigins    map[string]ast.Position
	constOrigins   map[string]ast.Position
	patternOrigins map[string]ast.Position
}

// newSymbolTable creates a new empty symbol table.
func newSymbolTable() *symbolTable {
	return &symbolTable{
		types:          make(map[string]*TypeSymbol),
		enums:          make(map[string]*EnumSymbol),
		consts:         make(map[string]*ConstSymbol),
		patterns:       make(map[string]*PatternSymbol),
		rpcs:           make(map[string]*RPCSymbol),
		standaloneDocs: []*DocSymbol{},
		typeOrigins:    make(map[string]ast.Position),
		enumOrigins:    make(map[string]ast.Position),
		constOrigins:   make(map[string]ast.Position),
		patternOrigins: make(map[string]ast.Position),
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

// registerPattern attempts to register a pattern symbol.
// Returns a Diagnostic if the pattern name is already declared.
func (st *symbolTable) registerPattern(sym *PatternSymbol) *Diagnostic {
	if existing, ok := st.patternOrigins[sym.Name]; ok {
		diag := newDiagnostic(
			sym.File,
			sym.Pos,
			sym.EndPos,
			CodeDuplicatePattern,
			formatDuplicateError("pattern", sym.Name, existing),
		)
		return &diag
	}
	st.patterns[sym.Name] = sym
	st.patternOrigins[sym.Name] = sym.Pos
	return nil
}

// registerRPC registers or merges an RPC symbol.
// RPCs with the same name are merged together.
func (st *symbolTable) registerRPC(sym *RPCSymbol) {
	if existing, ok := st.rpcs[sym.Name]; ok {
		// Merge: add the new file to DeclaredIn
		existing.DeclaredIn = append(existing.DeclaredIn, sym.DeclaredIn...)
		// Merge procs (individual duplicates are checked later)
		for name, proc := range sym.Procs {
			existing.Procs[name] = proc
		}
		// Merge streams
		for name, stream := range sym.Streams {
			existing.Streams[name] = stream
		}
		// Merge standalone docs
		existing.StandaloneDocs = append(existing.StandaloneDocs, sym.StandaloneDocs...)
	} else {
		st.rpcs[sym.Name] = sym
	}
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

// lookupConst returns the const symbol with the given name, or nil if not found.
func (st *symbolTable) lookupConst(name string) *ConstSymbol {
	return st.consts[name]
}

// lookupPattern returns the pattern symbol with the given name, or nil if not found.
func (st *symbolTable) lookupPattern(name string) *PatternSymbol {
	return st.patterns[name]
}

// lookupRPC returns the RPC symbol with the given name, or nil if not found.
func (st *symbolTable) lookupRPC(name string) *RPCSymbol {
	return st.rpcs[name]
}

// typeExists checks if a type with the given name exists.
// This includes both custom types and primitive types.
func (st *symbolTable) typeExists(name string) bool {
	if ast.IsPrimitiveType(name) {
		return true
	}
	_, ok := st.types[name]
	if ok {
		return true
	}
	// Also check enums as valid types
	_, ok = st.enums[name]
	return ok
}

// buildProgram creates a Program from the collected symbols.
func (st *symbolTable) buildProgram(entryPoint string, files map[string]*File) *Program {
	return &Program{
		EntryPoint:     entryPoint,
		Files:          files,
		Types:          st.types,
		Enums:          st.enums,
		Consts:         st.consts,
		Patterns:       st.patterns,
		RPCs:           st.rpcs,
		StandaloneDocs: st.standaloneDocs,
	}
}

// formatDuplicateError creates a standardized duplicate declaration error message.
func formatDuplicateError(kind, name string, original ast.Position) string {
	return fmt.Sprintf("%s %q is already declared at %s:%d:%d",
		kind, name, original.Filename, original.Line, original.Column)
}
