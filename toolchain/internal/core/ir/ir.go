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

//go:generate go run ir_gen.go

import (
	"encoding/json"
)

// ============================================================================
// SCHEMA ROOT
// ============================================================================

// Schema is the root of the IR - the "Golden Platter" for generators.
// All collections are sorted alphabetically for deterministic output.
type Schema struct {
	Types          []Type     `json:"types"`
	Enums          []Enum     `json:"enums"`
	Constants      []Constant `json:"constants"`
	Patterns       []Pattern  `json:"patterns"`
	RPCs           []RPC      `json:"rpcs"`
	StandaloneDocs []string   `json:"standaloneDocs,omitempty"`
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
	Name       string       `json:"name"`
	Doc        string       `json:"doc,omitempty"`
	Deprecated *Deprecation `json:"deprecated,omitempty"`
	Fields     []Field      `json:"fields"`
}

// Field represents a field with its type fully resolved.
type Field struct {
	Name     string  `json:"name"`
	Doc      string  `json:"doc,omitempty"`
	Optional bool    `json:"optional,omitempty"`
	Type     TypeRef `json:"type"`
}

// TypeRef represents any type in a unified way.
// Only the fields relevant to the Kind are populated.
type TypeRef struct {
	Kind            TypeKind      `json:"kind"`
	Primitive       PrimitiveType `json:"primitive,omitempty"`       // If Kind == "primitive"
	Type            string        `json:"type,omitempty"`            // If Kind == "type" (custom type name)
	Enum            string        `json:"enum,omitempty"`            // If Kind == "enum" (enum name)
	EnumInfo        *EnumInfo     `json:"enumInfo,omitempty"`        // If Kind == "enum"
	ArrayItem       *TypeRef      `json:"arrayItem,omitempty"`       // If Kind == "array" - the base element type
	ArrayDimensions int           `json:"arrayDimensions,omitempty"` // If Kind == "array" - number of array dimensions (e.g., 2 for int[][])
	MapValue        *TypeRef      `json:"mapValue,omitempty"`        // If Kind == "map"
	Object          *InlineObject `json:"object,omitempty"`          // If Kind == "object"
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
	ValueType EnumValueType `json:"valueType"`
}

// InlineObject represents an anonymous inline object type.
type InlineObject struct {
	Fields []Field `json:"fields"`
}

// ============================================================================
// ENUMS
// ============================================================================

// Enum represents an enumeration type.
type Enum struct {
	Name       string        `json:"name"`
	Doc        string        `json:"doc,omitempty"`
	Deprecated *Deprecation  `json:"deprecated,omitempty"`
	ValueType  EnumValueType `json:"valueType"`
	Members    []EnumMember  `json:"members"`
}

// EnumValueType indicates whether an enum uses string or integer values.
type EnumValueType string

const (
	EnumValueTypeString EnumValueType = "string"
	EnumValueTypeInt    EnumValueType = "int"
)

// EnumMember represents a member of an enum.
type EnumMember struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ============================================================================
// CONSTANTS
// ============================================================================

// Constant represents a constant declaration.
type Constant struct {
	Name       string         `json:"name"`
	Doc        string         `json:"doc,omitempty"`
	Deprecated *Deprecation   `json:"deprecated,omitempty"`
	ValueType  ConstValueType `json:"valueType"`
	Value      string         `json:"value"`
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
	Name         string       `json:"name"`
	Doc          string       `json:"doc,omitempty"`
	Deprecated   *Deprecation `json:"deprecated,omitempty"`
	Template     string       `json:"template"`
	Placeholders []string     `json:"placeholders"`
}

// ============================================================================
// RPC SERVICES
// ============================================================================

// RPC represents an RPC service containing procedures and streams.
type RPC struct {
	Name       string       `json:"name"`
	Doc        string       `json:"doc,omitempty"`
	Deprecated *Deprecation `json:"deprecated,omitempty"`
	Procs      []Procedure  `json:"procs"`
	Streams    []Stream     `json:"streams"`
	Docs       []string     `json:"docs,omitempty"` // Standalone docs within the RPC
}

// Procedure represents an RPC procedure (request-response).
type Procedure struct {
	Name       string       `json:"name"`
	Doc        string       `json:"doc,omitempty"`
	Deprecated *Deprecation `json:"deprecated,omitempty"`
	Input      []Field      `json:"input"`
	Output     []Field      `json:"output"`
}

// Stream represents an RPC stream (server-sent events).
type Stream struct {
	Name       string       `json:"name"`
	Doc        string       `json:"doc,omitempty"`
	Deprecated *Deprecation `json:"deprecated,omitempty"`
	Input      []Field      `json:"input"`
	Output     []Field      `json:"output"`
}

// ============================================================================
// SHARED
// ============================================================================

// Deprecation contains information about a deprecated element.
type Deprecation struct {
	Message string `json:"message,omitempty"`
}
