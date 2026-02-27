package ast

import (
	"slices"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/util/strutil"
)

// QuotedString is a custom type that implements participle's Capture interface
// to automatically strip surrounding double quotes from StringLiteral tokens.
type QuotedString string

// Capture implements the participle Capture interface.
func (q *QuotedString) Capture(values []string) error {
	s := values[0]
	// Strip surrounding quotes if present
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	*q = QuotedString(s)
	return nil
}

// String returns the underlying string value.
func (q QuotedString) String() string {
	return string(q)
}

// DocstringValue is a custom type that implements participle's Capture interface
// to automatically strip surrounding triple-quote delimiters from Docstring tokens.
type DocstringValue string

// Capture implements the participle Capture interface.
func (d *DocstringValue) Capture(values []string) error {
	s := values[0]
	// Strip surrounding """ if present
	if len(s) >= 6 && strings.HasPrefix(s, `"""`) && strings.HasSuffix(s, `"""`) {
		s = s[3 : len(s)-3]
	}
	*d = DocstringValue(s)
	return nil
}

// String returns the underlying string value.
func (d DocstringValue) String() string {
	return string(d)
}

// This AST is used for parsing the schema and it uses the
// participle library for parsing.
//
// It includes embedded Positions fields for each node to track the
// position of the node in the original source code, it is used
// later in the analyzer and LSP to give useful error messages
// and auto-completion. Those positions are automatically populated
// by the participle library.

// PrimitiveType represents a primitive type.
type PrimitiveType = string

// PrimitiveType constants.
const (
	PrimitiveTypeString   PrimitiveType = "string"
	PrimitiveTypeInt      PrimitiveType = "int"
	PrimitiveTypeFloat    PrimitiveType = "float"
	PrimitiveTypeBool     PrimitiveType = "bool"
	PrimitiveTypeDatetime PrimitiveType = "datetime"
)

// PrimitiveTypes is a list of primitive types that are not
// considered as custom types.
var PrimitiveTypes = []PrimitiveType{
	PrimitiveTypeString,
	PrimitiveTypeInt,
	PrimitiveTypeFloat,
	PrimitiveTypeBool,
	PrimitiveTypeDatetime,
}

// IsPrimitiveType checks if a type is a primitive type.
func IsPrimitiveType(name PrimitiveType) bool {
	return slices.Contains(PrimitiveTypes, name)
}

// Schema is the root of the schema AST.
type Schema struct {
	Positions
	Declarations []*TopLevelDecl `parser:"@@*"`
}

// GetIncludes returns all include declarations in the schema.
func (s *Schema) GetIncludes() []*Include {
	includes := []*Include{}
	for _, node := range s.Declarations {
		if node.Kind() == DeclKindInclude {
			includes = append(includes, node.Include)
		}
	}
	return includes
}

// GetDocstrings returns all docstrings in the schema.
func (s *Schema) GetDocstrings() []*Docstring {
	docstrings := []*Docstring{}
	for _, node := range s.Declarations {
		if node.Kind() == DeclKindDocstring {
			docstrings = append(docstrings, node.Docstring)
		}
	}
	return docstrings
}

// GetTypes returns all custom types in the schema.
func (s *Schema) GetTypes() []*TypeDecl {
	types := []*TypeDecl{}
	for _, node := range s.Declarations {
		if node.Kind() == DeclKindType {
			types = append(types, node.Type)
		}
	}
	return types
}

// GetTypesMap returns a map of type names to type declarations.
func (s *Schema) GetTypesMap() map[string]*TypeDecl {
	typesMap := make(map[string]*TypeDecl)
	for _, typeDecl := range s.GetTypes() {
		typesMap[typeDecl.Name] = typeDecl
	}
	return typesMap
}

// GetConsts returns all constant declarations in the schema.
func (s *Schema) GetConsts() []*ConstDecl {
	consts := []*ConstDecl{}
	for _, node := range s.Declarations {
		if node.Kind() == DeclKindConst {
			consts = append(consts, node.Const)
		}
	}
	return consts
}

// GetEnums returns all enum declarations in the schema.
func (s *Schema) GetEnums() []*EnumDecl {
	enums := []*EnumDecl{}
	for _, node := range s.Declarations {
		if node.Kind() == DeclKindEnum {
			enums = append(enums, node.Enum)
		}
	}
	return enums
}

// DeclKind represents the kind of a schema child node.
type DeclKind string

const (
	DeclKindInclude   DeclKind = "Include"
	DeclKindDocstring DeclKind = "Docstring"
	DeclKindType      DeclKind = "Type"
	DeclKindConst     DeclKind = "Const"
	DeclKindEnum      DeclKind = "Enum"
)

// TopLevelDecl represents a child node of the Schema root node.
type TopLevelDecl struct {
	Positions
	Include   *Include   `parser:"  @@"`
	Type      *TypeDecl  `parser:"| @@"`
	Const     *ConstDecl `parser:"| @@"`
	Enum      *EnumDecl  `parser:"| @@"`
	Docstring *Docstring `parser:"| @@"`
}

func (n *TopLevelDecl) Kind() DeclKind {
	if n.Include != nil {
		return DeclKindInclude
	}
	if n.Docstring != nil {
		return DeclKindDocstring
	}
	if n.Type != nil {
		return DeclKindType
	}
	if n.Const != nil {
		return DeclKindConst
	}
	if n.Enum != nil {
		return DeclKindEnum
	}
	return ""
}

// Include represents an include statement.
type Include struct {
	Positions
	Path QuotedString `parser:"Include @StringLiteral"`
}

// Annotation represents metadata attached to declarations and fields.
type Annotation struct {
	Positions
	Name     string       `parser:"At @Ident"`
	Argument *DataLiteral `parser:"(LParen @@ RParen)?"`
}

// TypeDecl represents a custom type declaration.
type TypeDecl struct {
	Positions
	Docstring   *Docstring    `parser:"(@@ (?! Newline Newline))?"`
	Annotations []*Annotation `parser:"@@*"`
	Name        string        `parser:"Type @Ident"`
	Members     []*TypeMember `parser:"LBrace @@* RBrace"`
}

// TypeMember represents a member within a type declaration block.
// Can be a Field, a Spread, or a standalone Docstring.
type TypeMember struct {
	Positions
	Field     *Field     `parser:"  @@"`
	Spread    *Spread    `parser:"| @@"`
	Docstring *Docstring `parser:"| @@"`
}

// GetFlattenedFields returns a recursive flattened list of all fields in the type declaration.
func (t *TypeDecl) GetFlattenedFields() []*Field {
	fields := []*Field{}
	for _, child := range t.Members {
		if child.Field == nil {
			continue
		}
		fields = append(fields, child.Field.GetFlattenedField()...)
	}
	return fields
}

// ConstDecl represents a constant declaration.
type ConstDecl struct {
	Positions
	Docstring   *Docstring    `parser:"(@@ (?! Newline Newline))?"`
	Annotations []*Annotation `parser:"@@*"`
	Name        string        `parser:"Const @Ident"`
	TypeName    *string       `parser:"(@Ident)?"`
	Value       *DataLiteral  `parser:"Equals @@"`
}

// EnumDecl represents an enum declaration.
type EnumDecl struct {
	Positions
	Docstring   *Docstring    `parser:"(@@ (?! Newline Newline))?"`
	Annotations []*Annotation `parser:"@@*"`
	Name        string        `parser:"Enum @Ident"`
	Members     []*EnumMember `parser:"LBrace @@* RBrace"`
}

// EnumMember represents a member of an enum.
// The Spread alternative is tried first (matching "...TypeName").
// If that fails, the parser falls through to match an annotatable named member
// with optional docstring, annotations, name, and value.
type EnumMember struct {
	Positions
	Spread      *Spread       `parser:"  @@"`
	Docstring   *Docstring    `parser:"| (@@ (?! Newline Newline))?"`
	Annotations []*Annotation `parser:"  @@*"`
	Name        string        `parser:"  @(Ident | Include | Const | Enum | Map | Type | String | Int | Float | Bool | Datetime | True | False)"`
	Value       *EnumValue    `parser:"  (Equals @@)?"`
}

// EnumValue represents the value of an enum member.
type EnumValue struct {
	Positions
	Str *QuotedString `parser:"  @StringLiteral"`
	Int *string       `parser:"| @IntLiteral"`
}

// Docstring represents a docstring in the schema.
type Docstring struct {
	Positions
	Value DocstringValue `parser:"@Docstring"`
}

// GetExternal returns a path and a bool indicating if the docstring
// references an external Markdown file.
func (d Docstring) GetExternal() (string, bool) {
	return DocstringIsExternal(string(d.Value))
}

// DocstringIsExternal checks if a docstring is an external markdown file.
//
// If it is, it returns the trimmed docstring and true.
// If it is not, it returns an empty string and false.
func DocstringIsExternal(docstring string) (string, bool) {
	trimmed := strings.TrimSpace(docstring)
	if strings.ContainsAny(trimmed, "\r\n") {
		return "", false
	}

	if strings.TrimSuffix(".md", trimmed) == "" {
		return "", false
	}

	if !strings.HasSuffix(trimmed, ".md") {
		return "", false
	}

	return trimmed, true
}

// Spread represents a spread operator (...Name or ...Namespace.Name).
type Spread struct {
	Positions
	Ref *Reference `parser:"Spread @@"`
}

// Field represents a field definition.
type Field struct {
	Positions
	Docstring   *Docstring    `parser:"(@@ (?! Newline Newline))?"`
	Annotations []*Annotation `parser:"@@*"`
	Name        string        `parser:"@(Ident | Include | Const | Enum | Map | Type | String | Int | Float | Bool | Datetime | True | False)"`
	Optional    bool          `parser:"@Question?"`
	Type        FieldType     `parser:"@@"`
}

// GetFlattenedField returns a recursive flattened list of this field and all its children fields.
func (f *Field) GetFlattenedField() []*Field {
	fields := []*Field{f}

	if f.Type.Base == nil || f.Type.Base.Object == nil {
		return fields
	}

	for _, child := range f.Type.Base.Object.Members {
		if child.Field == nil {
			continue
		}
		fields = append(fields, child.Field.GetFlattenedField()...)
	}

	return fields
}

// ArrayDimensions represents the number of array dimensions.
// 0 means not an array, 1 means [], 2 means [][], etc.
type ArrayDimensions int

// Capture implements participle's Capture interface to count array bracket pairs.
// participle calls Capture once per match, so we accumulate instead of replacing.
func (a *ArrayDimensions) Capture(values []string) error {
	*a += ArrayDimensions(len(values))
	return nil
}

// FieldType represents the type of a field.
type FieldType struct {
	Positions
	Base       *FieldTypeBase  `parser:"@@"`
	Dimensions ArrayDimensions `parser:"(LBracket @RBracket)*"`
}

// IsArray returns true if this type has at least one array dimension.
func (ft *FieldType) IsArray() bool {
	return ft.Dimensions > 0
}

// FieldTypeBase represents the base type of a field (named, map, or inline object).
type FieldTypeBase struct {
	Positions
	// Named can be a primitive type (string, int, float, bool, datetime) or a custom type name
	Named  *string          `parser:"  @(Ident | String | Int | Float | Bool | Datetime)"`
	Map    *FieldTypeMap    `parser:"| @@"`
	Object *FieldTypeObject `parser:"| @@"`
}

// FieldTypeMap represents a map type: map<ValueType>.
type FieldTypeMap struct {
	Positions
	ValueType *FieldType `parser:"Map LBracket @@ RBracket"`
}

// FieldTypeObject represents an inline object type definition.
type FieldTypeObject struct {
	Positions
	Members []*TypeMember `parser:"LBrace @@* RBrace"`
}

// DataLiteral represents any data literal used by constants and annotations.
type DataLiteral struct {
	Positions
	Object *DataLiteralObject `parser:"  @@"`
	Array  *DataLiteralArray  `parser:"| @@"`
	Scalar *ScalarLiteral     `parser:"| @@"`
}

// DataLiteralObject represents an object literal with key/value entries and spreads.
type DataLiteralObject struct {
	Positions
	Entries []*DataLiteralObjectEntry `parser:"LBrace @@* RBrace"`
}

// DataLiteralObjectEntry represents one object literal entry.
type DataLiteralObjectEntry struct {
	Positions
	Spread *Spread      `parser:"  @@"`
	Key    string       `parser:"| @(Ident | Include | Const | Enum | Map | Type | String | Int | Float | Bool | Datetime | True | False)"`
	Value  *DataLiteral `parser:"@@"`
}

// DataLiteralArray represents an array literal.
type DataLiteralArray struct {
	Positions
	Elements []*DataLiteral `parser:"LBracket @@* RBracket"`
}

// ScalarLiteral represents any of the built-in literal types or a reference.
type ScalarLiteral struct {
	Positions
	Str   *QuotedString `parser:"  @StringLiteral"`
	Float *string       `parser:"| @FloatLiteral"`
	Int   *string       `parser:"| @IntLiteral"`
	True  bool          `parser:"| @True"`
	False bool          `parser:"| @False"`
	Ref   *Reference    `parser:"| @@"`
}

// String returns the string representation of the value of the literal.
func (al ScalarLiteral) String() string {
	if al.Str != nil {
		return `"` + strutil.EscapeQuotes(string(*al.Str)) + `"`
	}
	if al.Float != nil {
		return *al.Float
	}
	if al.Int != nil {
		return *al.Int
	}
	if al.True {
		return "true"
	}
	if al.False {
		return "false"
	}
	if al.Ref != nil {
		return al.Ref.String()
	}
	return ""
}

// Reference represents a reference to a type, constant or an enum member.
// Examples: FOO (const ref), Color.Red (enum member ref).
type Reference struct {
	Positions
	Name   string  `parser:"@Ident"`
	Member *string `parser:"(Dot @Ident)?"`
}

// String returns the string representation of the reference.
func (r Reference) String() string {
	if r.Member != nil {
		return r.Name + "." + *r.Member
	}
	return r.Name
}
