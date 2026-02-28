package analysis

import (
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
)

// Program represents a fully analyzed and validated VDL project.
// It contains all symbols merged into a global namespace with origin metadata
// for LSP support (go-to-definition, find-references, etc.).
type Program struct {
	// EntryPoint is the absolute path to the main .vdl file.
	EntryPoint string

	// Files contains all parsed files in the project, keyed by absolute path.
	Files map[string]*File

	// Global namespace - all symbols merged from all files
	Types          map[string]*TypeSymbol
	Enums          map[string]*EnumSymbol
	Consts         map[string]*ConstSymbol
	StandaloneDocs []*DocSymbol
}

// File represents a single parsed .vdl file in the project.
type File struct {
	Path     string      // Absolute path to the file
	AST      *ast.Schema // Parsed AST
	Includes []string    // Resolved absolute paths of included files
}

// Symbol contains common metadata for all symbol types.
// It provides origin information for LSP features.
type Symbol struct {
	Name        string       // The symbol's identifier
	File        string       // File where the symbol was declared
	Pos         ast.Position // Start position of the declaration
	EndPos      ast.Position // End position of the declaration
	Docstring   *string      // Resolved docstring content (nil if none)
	Annotations []*AnnotationRef
}

// AnnotationRef represents an annotation attached to a symbol.
type AnnotationRef struct {
	Name     string
	Argument *ast.DataLiteral
	Pos      ast.Position
	EndPos   ast.Position
}

// TypeSymbol represents a type declaration in the global namespace.
type TypeSymbol struct {
	Symbol
	AST     *ast.TypeDecl  // Original AST node
	Fields  []*FieldSymbol // Direct fields (not expanded from spreads)
	Spreads []*SpreadRef   // Spread references for validation
}

// FieldSymbol represents a field within a type or inline object.
type FieldSymbol struct {
	Symbol
	AST      *ast.Field     // Original AST node
	Optional bool           // Whether the field is optional (?)
	Type     *FieldTypeInfo // Type information
}

// FieldTypeInfo describes the type of a field.
type FieldTypeInfo struct {
	Kind      FieldTypeKind  // The kind of type
	Name      string         // Type name (for Primitive/Custom kinds)
	ArrayDims int            // Number of array dimensions (0 = not array)
	MapValue  *FieldTypeInfo // Value type for Map kinds
	ObjectDef *InlineObject  // Definition for Object kinds

	// ResolvedSymbol is the resolved type/enum symbol for Custom kinds.
	// This enables O(1) "Go to Definition" in LSP without re-lookup.
	// Only populated after validation; nil for primitives and unresolved types.
	ResolvedType *TypeSymbol `json:"-"`
	ResolvedEnum *EnumSymbol `json:"-"`
}

// FieldTypeKind indicates the category of a field type.
type FieldTypeKind int

const (
	FieldTypeKindPrimitive FieldTypeKind = iota // string, int, float, bool, datetime
	FieldTypeKindCustom                         // Reference to a custom type
	FieldTypeKindMap                            // map<ValueType>
	FieldTypeKindObject                         // Inline object { ... }
)

// SpreadRef represents a reference to a type via the spread operator.
type SpreadRef struct {
	Name   string       // Name part of the spread reference
	Member *string      // Optional member part (invalid for spread semantics)
	Pos    ast.Position // Position of the spread in source
	EndPos ast.Position
}

// InlineObject represents an inline object type definition.
type InlineObject struct {
	Fields  []*FieldSymbol // Fields in the inline object
	Spreads []*SpreadRef   // Spreads in the inline object
}

// EnumSymbol represents an enum declaration in the global namespace.
type EnumSymbol struct {
	Symbol
	AST       *ast.EnumDecl       // Original AST node
	ValueType EnumValueType       // Whether this is a string or int enum
	Members   []*EnumMemberSymbol // Enum members
	Spreads   []*SpreadRef        // Enum spreads
}

// EnumValueType indicates whether an enum uses string or integer values.
type EnumValueType int

const (
	EnumValueTypeString EnumValueType = iota // Default: member name as value
	EnumValueTypeInt                         // Explicit integer values
)

// EnumMemberSymbol represents a member of an enum.
type EnumMemberSymbol struct {
	Symbol
	Value       string // String representation of the value
	HasExplicit bool   // Whether value was explicitly set
}

// ConstSymbol represents a constant declaration in the global namespace.
type ConstSymbol struct {
	Symbol
	AST              *ast.ConstDecl // Original AST node
	ExplicitTypeName *string
	ValueType        ConstValueType // Inferred top-level value type
	Value            string         // String representation for scalar values
}

// ConstValueType indicates the type of a constant's value.
type ConstValueType int

const (
	ConstValueTypeString ConstValueType = iota
	ConstValueTypeInt
	ConstValueTypeFloat
	ConstValueTypeBool
	ConstValueTypeObject
	ConstValueTypeArray
	ConstValueTypeReference
	ConstValueTypeUnknown
)

// DocSymbol represents a standalone docstring.
type DocSymbol struct {
	Content string       // The resolved docstring content
	Pos     ast.Position // Position in source
	EndPos  ast.Position
	File    string // File where the docstring appears
}

// newProgram creates an empty Program ready for population.
func newProgram(entryPoint string) *Program {
	return &Program{
		EntryPoint:     entryPoint,
		Files:          make(map[string]*File),
		Types:          make(map[string]*TypeSymbol),
		Enums:          make(map[string]*EnumSymbol),
		Consts:         make(map[string]*ConstSymbol),
		StandaloneDocs: []*DocSymbol{},
	}
}
