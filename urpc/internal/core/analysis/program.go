package analysis

import (
	"github.com/uforg/uforpc/urpc/internal/core/ast"
)

// Program represents a fully analyzed and validated UFO RPC project.
// It contains all symbols merged into a global namespace with origin metadata
// for LSP support (go-to-definition, find-references, etc.).
type Program struct {
	// EntryPoint is the absolute path to the main .ufo file.
	EntryPoint string

	// Files contains all parsed files in the project, keyed by absolute path.
	Files map[string]*File

	// Global namespace - all symbols merged from all files
	Types    map[string]*TypeSymbol
	Enums    map[string]*EnumSymbol
	Consts   map[string]*ConstSymbol
	Patterns map[string]*PatternSymbol
	RPCs     map[string]*RPCSymbol

	// StandaloneDocs contains top-level standalone docstrings.
	StandaloneDocs []*DocSymbol
}

// File represents a single parsed .ufo file in the project.
type File struct {
	Path     string      // Absolute path to the file
	AST      *ast.Schema // Parsed AST
	Includes []string    // Resolved absolute paths of included files
}

// Symbol contains common metadata for all symbol types.
// It provides origin information for LSP features.
type Symbol struct {
	Name       string           // The symbol's identifier
	File       string           // File where the symbol was declared
	Pos        ast.Position     // Start position of the declaration
	EndPos     ast.Position     // End position of the declaration
	Docstring  *string          // Resolved docstring content (nil if none)
	Deprecated *DeprecationInfo // Deprecation info (nil if not deprecated)
}

// DeprecationInfo contains information about a deprecated symbol.
type DeprecationInfo struct {
	Message string // Optional deprecation message
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
	TypeName string       // The name of the type being spread
	Pos      ast.Position // Position of the spread in source
	EndPos   ast.Position
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
}

// EnumValueType indicates whether an enum uses string or integer values.
type EnumValueType int

const (
	EnumValueTypeString EnumValueType = iota // Default: member name as value
	EnumValueTypeInt                         // Explicit integer values
)

// EnumMemberSymbol represents a member of an enum.
type EnumMemberSymbol struct {
	Name        string       // Member name
	Pos         ast.Position // Position in source
	EndPos      ast.Position
	Value       string // String representation of the value
	HasExplicit bool   // Whether value was explicitly set
}

// ConstSymbol represents a constant declaration in the global namespace.
type ConstSymbol struct {
	Symbol
	AST       *ast.ConstDecl // Original AST node
	ValueType ConstValueType // The type of the constant value
	Value     string         // String representation of the value
}

// ConstValueType indicates the type of a constant's value.
type ConstValueType int

const (
	ConstValueTypeString ConstValueType = iota
	ConstValueTypeInt
	ConstValueTypeFloat
	ConstValueTypeBool
)

// PatternSymbol represents a pattern declaration in the global namespace.
type PatternSymbol struct {
	Symbol
	AST          *ast.PatternDecl // Original AST node
	Template     string           // The template string
	Placeholders []string         // Extracted placeholder names
}

// RPCSymbol represents an RPC service in the global namespace.
// Multiple RPC declarations with the same name are merged into one.
type RPCSymbol struct {
	Symbol
	Procs          map[string]*ProcSymbol   // Procedures in this RPC
	Streams        map[string]*StreamSymbol // Streams in this RPC
	DeclaredIn     []string                 // Files where this RPC was declared
	StandaloneDocs []*DocSymbol             // Standalone docs within the RPC
}

// ProcSymbol represents a procedure within an RPC service.
type ProcSymbol struct {
	Symbol
	AST    *ast.ProcDecl // Original AST node
	Input  *BlockSymbol  // Input block (nil if not defined)
	Output *BlockSymbol  // Output block (nil if not defined)
}

// StreamSymbol represents a stream within an RPC service.
type StreamSymbol struct {
	Symbol
	AST    *ast.StreamDecl // Original AST node
	Input  *BlockSymbol    // Input block (nil if not defined)
	Output *BlockSymbol    // Output block (nil if not defined)
}

// BlockSymbol represents an input or output block.
type BlockSymbol struct {
	Pos     ast.Position // Position of the block
	EndPos  ast.Position
	Fields  []*FieldSymbol // Fields in the block
	Spreads []*SpreadRef   // Spreads in the block
}

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
		Patterns:       make(map[string]*PatternSymbol),
		RPCs:           make(map[string]*RPCSymbol),
		StandaloneDocs: []*DocSymbol{},
	}
}
