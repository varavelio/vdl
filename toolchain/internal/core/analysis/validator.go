package analysis

import (
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// validator orchestrates symbol collection and validation.
type validator struct {
	symbols     *symbolTable
	files       map[string]*File
	diagnostics []Diagnostic
}

// newValidator creates a new validator instance.
func newValidator(files map[string]*File) *validator {
	return &validator{
		symbols:     newSymbolTable(),
		files:       files,
		diagnostics: []Diagnostic{},
	}
}

// collect processes all files and collects symbols into the symbol table.
// Returns any errors encountered during collection (e.g., duplicate declarations).
func (v *validator) collect() []Diagnostic {
	for _, file := range v.files {
		v.collectFromSchema(file.AST, file.Path)
	}
	return v.diagnostics
}

// collectFromSchema extracts all symbols from a single schema.
func (v *validator) collectFromSchema(schema *ast.Schema, file string) {
	// Collect standalone docstrings
	for _, doc := range schema.GetDocstrings() {
		v.symbols.addStandaloneDoc(&DocSymbol{
			Content: string(doc.Value),
			Pos:     doc.Pos,
			EndPos:  doc.EndPos,
			File:    file,
		})
	}

	// Collect types
	for _, typeDecl := range schema.GetTypes() {
		sym := v.buildTypeSymbol(typeDecl, file)
		if diag := v.symbols.registerType(sym); diag != nil {
			v.diagnostics = append(v.diagnostics, *diag)
		}
	}

	// Collect enums
	for _, enumDecl := range schema.GetEnums() {
		sym := buildEnumSymbol(enumDecl, file)
		if diag := v.symbols.registerEnum(sym); diag != nil {
			v.diagnostics = append(v.diagnostics, *diag)
		}
	}

	// Collect constants
	for _, constDecl := range schema.GetConsts() {
		sym := v.buildConstSymbol(constDecl, file)
		if diag := v.symbols.registerConst(sym); diag != nil {
			v.diagnostics = append(v.diagnostics, *diag)
		}
	}

	// Collect patterns
	for _, patternDecl := range schema.GetPatterns() {
		sym := buildPatternSymbol(patternDecl, file)
		if diag := v.symbols.registerPattern(sym); diag != nil {
			v.diagnostics = append(v.diagnostics, *diag)
		}
	}

	// Collect RPCs (these are merged, not duplicate-checked at schema level)
	for _, rpcDecl := range schema.GetRPCs() {
		sym, rpcDiags := buildRPCSymbol(rpcDecl, file)
		v.diagnostics = append(v.diagnostics, rpcDiags...)
		v.symbols.registerRPC(sym)
	}
}

// buildTypeSymbol creates a TypeSymbol from an AST TypeDecl.
func (v *validator) buildTypeSymbol(decl *ast.TypeDecl, file string) *TypeSymbol {
	var docstring *string
	if decl.Docstring != nil {
		s := string(decl.Docstring.Value)
		docstring = &s
	}

	var deprecated *DeprecationInfo
	if decl.Deprecated != nil {
		msg := ""
		if decl.Deprecated.Message != nil {
			msg = string(*decl.Deprecated.Message)
		}
		deprecated = &DeprecationInfo{Message: msg}
	}

	typ := &TypeSymbol{
		Symbol: Symbol{
			Name:       decl.Name,
			File:       file,
			Pos:        decl.Pos,
			EndPos:     decl.EndPos,
			Docstring:  docstring,
			Deprecated: deprecated,
		},
		AST:     decl,
		Fields:  []*FieldSymbol{},
		Spreads: []*SpreadRef{},
	}

	// Process children
	for _, child := range decl.Children {
		if child.Field != nil {
			typ.Fields = append(typ.Fields, buildFieldSymbol(child.Field, file))
		}
		if child.Spread != nil {
			typ.Spreads = append(typ.Spreads, &SpreadRef{
				TypeName: child.Spread.TypeName,
				Pos:      child.Spread.Pos,
				EndPos:   child.Spread.EndPos,
			})
		}
	}

	return typ
}

// buildConstSymbol creates a ConstSymbol from an AST ConstDecl.
func (v *validator) buildConstSymbol(decl *ast.ConstDecl, file string) *ConstSymbol {
	var docstring *string
	if decl.Docstring != nil {
		s := string(decl.Docstring.Value)
		docstring = &s
	}

	var deprecated *DeprecationInfo
	if decl.Deprecated != nil {
		msg := ""
		if decl.Deprecated.Message != nil {
			msg = string(*decl.Deprecated.Message)
		}
		deprecated = &DeprecationInfo{Message: msg}
	}

	cnst := &ConstSymbol{
		Symbol: Symbol{
			Name:       decl.Name,
			File:       file,
			Pos:        decl.Pos,
			EndPos:     decl.EndPos,
			Docstring:  docstring,
			Deprecated: deprecated,
		},
		AST: decl,
	}

	// Determine value type and value
	if decl.Value != nil {
		if decl.Value.Str != nil {
			cnst.ValueType = ConstValueTypeString
			cnst.Value = string(*decl.Value.Str)
		} else if decl.Value.Int != nil {
			cnst.ValueType = ConstValueTypeInt
			cnst.Value = *decl.Value.Int
		} else if decl.Value.Float != nil {
			cnst.ValueType = ConstValueTypeFloat
			cnst.Value = *decl.Value.Float
		} else if decl.Value.True {
			cnst.ValueType = ConstValueTypeBool
			cnst.Value = "true"
		} else if decl.Value.False {
			cnst.ValueType = ConstValueTypeBool
			cnst.Value = "false"
		}
	}

	return cnst
}

// validate runs all validation phases and returns collected diagnostics.
func (v *validator) validate() []Diagnostic {
	var diagnostics []Diagnostic

	// Run all validators
	diagnostics = append(diagnostics, validateNaming(v.symbols)...)
	diagnostics = append(diagnostics, validateTypes(v.symbols)...)
	diagnostics = append(diagnostics, validateSpreads(v.symbols)...)
	diagnostics = append(diagnostics, validateEnums(v.symbols)...)
	diagnostics = append(diagnostics, validatePatterns(v.symbols)...)
	diagnostics = append(diagnostics, validateRPCs(v.symbols)...)
	diagnostics = append(diagnostics, validateCycles(v.symbols)...)
	diagnostics = append(diagnostics, validateStructure(v.symbols, v.files)...)
	diagnostics = append(diagnostics, validateGlobalUniqueness(v.symbols)...)

	return diagnostics
}

// buildProgram creates the final Program from collected symbols.
func (v *validator) buildProgram(entryPoint string) *Program {
	return v.symbols.buildProgram(entryPoint, v.files)
}
