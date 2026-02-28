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
}

// buildTypeSymbol creates a TypeSymbol from an AST TypeDecl.
func (v *validator) buildTypeSymbol(decl *ast.TypeDecl, file string) *TypeSymbol {
	var docstring *string
	if decl.Docstring != nil {
		s := string(decl.Docstring.Value)
		docstring = &s
	}

	typ := &TypeSymbol{
		Symbol: Symbol{
			Name:        decl.Name,
			File:        file,
			Pos:         decl.Pos,
			EndPos:      decl.EndPos,
			Docstring:   docstring,
			Annotations: buildAnnotationRefs(decl.Annotations),
		},
		AST:     decl,
		Fields:  []*FieldSymbol{},
		Spreads: []*SpreadRef{},
	}

	// Process members
	for _, child := range decl.Members {
		if child.Field != nil {
			typ.Fields = append(typ.Fields, buildFieldSymbol(child.Field, file))
		}
		if child.Spread != nil {
			typ.Spreads = append(typ.Spreads, &SpreadRef{
				Name:   child.Spread.Ref.Name,
				Member: child.Spread.Ref.Member,
				Pos:    child.Spread.Pos,
				EndPos: child.Spread.EndPos,
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

	cnst := &ConstSymbol{
		Symbol: Symbol{
			Name:        decl.Name,
			File:        file,
			Pos:         decl.Pos,
			EndPos:      decl.EndPos,
			Docstring:   docstring,
			Annotations: buildAnnotationRefs(decl.Annotations),
		},
		AST:              decl,
		ExplicitTypeName: decl.TypeName,
		ValueType:        ConstValueTypeUnknown,
	}

	if decl.Value != nil && decl.Value.Scalar != nil {
		s := decl.Value.Scalar
		switch {
		case s.Str != nil:
			cnst.ValueType = ConstValueTypeString
			cnst.Value = string(*s.Str)
		case s.Int != nil:
			cnst.ValueType = ConstValueTypeInt
			cnst.Value = *s.Int
		case s.Float != nil:
			cnst.ValueType = ConstValueTypeFloat
			cnst.Value = *s.Float
		case s.True || s.False:
			cnst.ValueType = ConstValueTypeBool
			if s.True {
				cnst.Value = "true"
			} else {
				cnst.Value = "false"
			}
		case s.Ref != nil:
			cnst.ValueType = ConstValueTypeReference
			cnst.Value = s.Ref.String()
		}
	} else if decl.Value != nil && decl.Value.Object != nil {
		cnst.ValueType = ConstValueTypeObject
	} else if decl.Value != nil && decl.Value.Array != nil {
		cnst.ValueType = ConstValueTypeArray
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
		validateConsts,
		validateSpreads,
		validateEnums,
		validateCycles,
		validateStructure,
		validateGlobalUniqueness,
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

func buildAnnotationRefs(annotations []*ast.Annotation) []*AnnotationRef {
	if len(annotations) == 0 {
		return nil
	}
	refs := make([]*AnnotationRef, 0, len(annotations))
	for _, ann := range annotations {
		refs = append(refs, &AnnotationRef{
			Name:     ann.Name,
			Argument: ann.Argument,
			Pos:      ann.Pos,
			EndPos:   ann.EndPos,
		})
	}
	return refs
}

// buildProgram creates the final Program from collected symbols.
func (v *validator) buildProgram(entryPoint string) *Program {
	return v.symbols.buildProgram(entryPoint, v.files)
}
