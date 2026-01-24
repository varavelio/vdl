// Package ir provides the Intermediate Representation for VDL schemas.
//
// The IR is the "Golden Platter" for code generators. It transforms the complex,
// validated analysis.Program into a clean, flat, serializable data structure
// optimized for code generation.
//
// Design principles:
//   - Total Source Amnesia: No line numbers, file paths, or AST references
//   - Aggressive Flattening: All spreads are expanded, generators see flat field lists
//   - Pre-Computed Documentation: All docs are inline strings, normalized and ready to use
//   - Linear & Deterministic: All collections are sorted slices, no maps
//   - Type Normalization: Unified TypeRef system for all type representations
//
// Usage:
//
//	program, diags := analysis.Analyze(fs, "main.vdl")
//	if len(diags) > 0 {
//	    // handle errors
//	}
//	schema := ir.FromProgram(program)
//	json, _ := schema.ToJSON()
package ir

import (
	"encoding/json"

	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// ============================================================================
// SCHEMA ROOT
// ============================================================================

// Schema is the root of the IR - the "Golden Platter" for generators.
// All collections are sorted alphabetically for deterministic output.
type Schema struct {
	Types      []Type      `json:"types" jsonschema:"description=List of all type definitions in the schema"`
	Enums      []Enum      `json:"enums" jsonschema:"description=List of all enum definitions in the schema"`
	Constants  []Constant  `json:"constants" jsonschema:"description=List of all constant definitions in the schema"`
	Patterns   []Pattern   `json:"patterns" jsonschema:"description=List of all pattern definitions in the schema"`
	RPCs       []RPC       `json:"rpcs" jsonschema:"description=List of all RPC service definitions in the schema"`
	Procedures []Procedure `json:"procedures" jsonschema:"description=Flattened list of all procedures from all RPCs"`
	Streams    []Stream    `json:"streams" jsonschema:"description=Flattened list of all streams from all RPCs"`
	Docs       []Doc       `json:"docs" jsonschema:"description=List of all standalone documentation blocks (even from all RPCs)"`
}

// ToJSON serializes the Schema to indented JSON.
func (s *Schema) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// ============================================================================
// TYPES
// ============================================================================

// Type represents a type with all fields expanded (no spreads visible).
type Type struct {
	Name       string       `json:"name" jsonschema:"description=The unique name of the type"`
	Doc        string       `json:"doc,omitempty" jsonschema:"description=Documentation for the type"`
	Deprecated *Deprecation `json:"deprecated,omitempty" jsonschema:"description=Deprecation status if deprecated"`
	Fields     []Field      `json:"fields" jsonschema:"description=List of fields in the type"`
}

// Field represents a field with its type fully resolved.
type Field struct {
	Name     string  `json:"name" jsonschema:"description=The name of the field"`
	Doc      string  `json:"doc,omitempty" jsonschema:"description=Documentation for the field"`
	Optional bool    `json:"optional,omitempty" jsonschema:"description=Whether the field is optional"`
	Type     TypeRef `json:"type" jsonschema:"description=The type definition of the field"`
}

// TypeRef represents any type in a unified way.
// Only the fields relevant to the Kind are populated.
type TypeRef struct {
	Kind            TypeKind      `json:"kind" jsonschema:"description=The category of the type (primitive\\, type\\, enum\\, array\\, map\\, object)"`
	Primitive       PrimitiveType `json:"primitive,omitempty" jsonschema:"description=The specific primitive type if Kind is primitive"`
	Type            string        `json:"type,omitempty" jsonschema:"description=The name of the custom type if Kind is type"`
	Enum            string        `json:"enum,omitempty" jsonschema:"description=The name of the enum if Kind is enum"`
	EnumInfo        *EnumInfo     `json:"enumInfo,omitempty" jsonschema:"description=Additional metadata for enum types"`
	ArrayItem       *TypeRef      `json:"arrayItem,omitempty" jsonschema:"description=The type of elements in the array if Kind is array"`
	ArrayDimensions int           `json:"arrayDimensions,omitempty" jsonschema:"description=Number of array dimensions (e.g. 2 for int[][])"`
	MapValue        *TypeRef      `json:"mapValue,omitempty" jsonschema:"description=The type of values in the map if Kind is map"`
	Object          *InlineObject `json:"object,omitempty" jsonschema:"description=Definition of the inline object if Kind is object"`
}

// TypeKind indicates the category of a type.
type TypeKind string

const (
	TypeKindPrimitive TypeKind = "primitive"
	TypeKindType      TypeKind = "type" // Custom type reference
	TypeKindEnum      TypeKind = "enum" // Enum reference
	TypeKindArray     TypeKind = "array"
	TypeKindMap       TypeKind = "map"
	TypeKindObject    TypeKind = "object"
)

// PrimitiveType represents built-in primitive types.
type PrimitiveType string

const (
	PrimitiveString   PrimitiveType = "string"
	PrimitiveInt      PrimitiveType = "int"
	PrimitiveFloat    PrimitiveType = "float"
	PrimitiveBool     PrimitiveType = "bool"
	PrimitiveDatetime PrimitiveType = "datetime"
)

// EnumInfo contains metadata about a referenced enum type.
type EnumInfo struct {
	ValueType EnumValueType `json:"valueType" jsonschema:"description=The underlying value type of the enum (string or int)"`
}

// InlineObject represents an anonymous inline object type.
type InlineObject struct {
	Fields []Field `json:"fields" jsonschema:"description=List of fields in the inline object"`
}

// ============================================================================
// ENUMS
// ============================================================================

// Enum represents an enumeration type.
type Enum struct {
	Name       string        `json:"name" jsonschema:"description=The unique name of the enum"`
	Doc        string        `json:"doc,omitempty" jsonschema:"description=Documentation for the enum"`
	Deprecated *Deprecation  `json:"deprecated,omitempty" jsonschema:"description=Deprecation status if deprecated"`
	ValueType  EnumValueType `json:"valueType" jsonschema:"description=The type of values contained in this enum"`
	Members    []EnumMember  `json:"members" jsonschema:"description=List of enum members"`
}

// EnumValueType indicates whether an enum uses string or integer values.
type EnumValueType string

const (
	EnumValueTypeString EnumValueType = "string"
	EnumValueTypeInt    EnumValueType = "int"
)

// EnumMember represents a member of an enum.
type EnumMember struct {
	Name  string `json:"name" jsonschema:"description=The name of the enum member"`
	Value string `json:"value" jsonschema:"description=The value of the enum member"`
}

// ============================================================================
// CONSTANTS
// ============================================================================

// Constant represents a constant declaration.
type Constant struct {
	Name       string         `json:"name" jsonschema:"description=The unique name of the constant"`
	Doc        string         `json:"doc,omitempty" jsonschema:"description=Documentation for the constant"`
	Deprecated *Deprecation   `json:"deprecated,omitempty" jsonschema:"description=Deprecation status if deprecated"`
	ValueType  ConstValueType `json:"valueType" jsonschema:"description=The type of the constant value"`
	Value      string         `json:"value" jsonschema:"description=The value of the constant as a string"`
}

// ConstValueType indicates the type of a constant's value.
type ConstValueType string

const (
	ConstValueTypeString ConstValueType = "string"
	ConstValueTypeInt    ConstValueType = "int"
	ConstValueTypeFloat  ConstValueType = "float"
	ConstValueTypeBool   ConstValueType = "bool"
)

// ============================================================================
// PATTERNS
// ============================================================================

// Pattern represents a pattern template for generating dynamic strings.
type Pattern struct {
	Name         string       `json:"name" jsonschema:"description=The unique name of the pattern"`
	Doc          string       `json:"doc,omitempty" jsonschema:"description=Documentation for the pattern"`
	Deprecated   *Deprecation `json:"deprecated,omitempty" jsonschema:"description=Deprecation status if deprecated"`
	Template     string       `json:"template" jsonschema:"description=The template string containing placeholders"`
	Placeholders []string     `json:"placeholders" jsonschema:"description=List of placeholder names extracted from the template"`
}

// ============================================================================
// RPC SERVICES
// ============================================================================

// RPC represents an RPC service containing procedures and streams.
type RPC struct {
	Name       string       `json:"name" jsonschema:"description=The unique name of the RPC service"`
	Doc        string       `json:"doc,omitempty" jsonschema:"description=Documentation for the RPC service"`
	Deprecated *Deprecation `json:"deprecated,omitempty" jsonschema:"description=Deprecation status if deprecated"`
	Procs      []Procedure  `json:"procs" jsonschema:"description=List of procedures in this RPC"`
	Streams    []Stream     `json:"streams" jsonschema:"description=List of streams in this RPC"`
	Docs       []string     `json:"docs,omitempty" jsonschema:"description=Standalone documentation blocks specific to this RPC"`
}

// Procedure represents an RPC procedure (request-response).
type Procedure struct {
	RPCName    string       `json:"rpcName" jsonschema:"description=The name of the parent RPC service"`
	Name       string       `json:"name" jsonschema:"description=The name of the procedure"`
	Doc        string       `json:"doc,omitempty" jsonschema:"description=Documentation for the procedure"`
	Deprecated *Deprecation `json:"deprecated,omitempty" jsonschema:"description=Deprecation status if deprecated"`
	Input      []Field      `json:"input" jsonschema:"description=List of input parameters"`
	Output     []Field      `json:"output" jsonschema:"description=List of output parameters"`
}

// FullName returns the fully qualified procedure name: {RPC}{Proc}
func (p Procedure) FullName() string {
	return p.RPCName + p.Name
}

// Path returns the RPC path for a procedure: {RpcName}/{ProcName}
func (p Procedure) Path() string {
	return strutil.ToPascalCase(p.RPCName) + "/" + strutil.ToPascalCase(p.Name)
}

// Stream represents an RPC stream (server-sent events).
type Stream struct {
	RPCName    string       `json:"rpcName" jsonschema:"description=The name of the parent RPC service"`
	Name       string       `json:"name" jsonschema:"description=The name of the stream"`
	Doc        string       `json:"doc,omitempty" jsonschema:"description=Documentation for the stream"`
	Deprecated *Deprecation `json:"deprecated,omitempty" jsonschema:"description=Deprecation status if deprecated"`
	Input      []Field      `json:"input" jsonschema:"description=List of input parameters"`
	Output     []Field      `json:"output" jsonschema:"description=List of output parameters"`
}

// FullName returns the fully qualified stream name: {RPC}{Stream}
func (s Stream) FullName() string {
	return s.RPCName + s.Name
}

// Path returns the RPC path for a stream: {RpcName}/{StreamName}
func (s Stream) Path() string {
	return strutil.ToPascalCase(s.RPCName) + "/" + strutil.ToPascalCase(s.Name)
}

// ============================================================================
// DOCUMENTATION
// ============================================================================

// Doc represents a standalone documentation block.
type Doc struct {
	RPCName string `json:"rpcName,omitempty" jsonschema:"description=The name of the RPC service this doc belongs to\\, if empty this is a global standalone documentation block"`
	Content string `json:"content" jsonschema:"description=The content of the documentation"`
}

// ============================================================================
// SHARED
// ============================================================================

// Deprecation contains information about a deprecated element.
type Deprecation struct {
	Message string `json:"message,omitempty" jsonschema:"description=A message explaining the deprecation and suggested alternative"`
}
