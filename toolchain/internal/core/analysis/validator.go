package analysis

import (
	"context"

	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// validator orchestrates symbol collection and validation.
type validator struct {
	ctx         context.Context
	symbols     *symbolTable
	files       map[string]*File
	diagnostics []Diagnostic
}

// newValidator creates a new validator instance without context (uses background context).
func newValidator(files map[string]*File) *validator {
	return newValidatorWithContext(context.Background(), files)
}

// newValidatorWithContext creates a new validator instance with context support for cancellation.
func newValidatorWithContext(ctx context.Context, files map[string]*File) *validator {
	return &validator{
		ctx:         ctx,
		symbols:     newSymbolTable(),
		files:       files,
		diagnostics: []Diagnostic{},
	}
}

// collect processes all files and collects symbols into the symbol table.
// Returns any errors encountered during collection (e.g., duplicate declarations).
// If the context is cancelled, returns partial results.
func (v *validator) collect() []Diagnostic {
	for _, file := range v.files {
		// Check for cancellation between files
		if v.ctx.Err() != nil {
			return v.diagnostics
		}
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
// If the context is cancelled, returns partial validation results.
func (v *validator) validate() []Diagnostic {
	var diagnostics []Diagnostic

	// List of validators to run
	validators := []func(*symbolTable) []Diagnostic{
		validateNaming,
		validateTypes,
		validateSpreads,
		validateEnums,
		validatePatterns,
		validateRPCs,
		validateCycles,
		func(s *symbolTable) []Diagnostic { return validateStructure(s, v.files) },
		validateGlobalUniqueness,
		validateCollisions,
	}

	// Run validators with cancellation checks between each
	for _, validate := range validators {
		if v.ctx.Err() != nil {
			return diagnostics
		}
		diagnostics = append(diagnostics, validate(v.symbols)...)
	}

	return diagnostics
}

// buildProgram creates the final Program from collected symbols.
func (v *validator) buildProgram(entryPoint string) *Program {
	return v.symbols.buildProgram(entryPoint, v.files)
}
