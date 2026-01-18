package ast

import (
	"slices"
	"strings"

	"github.com/varavelio/vdl/urpc/internal/util/strutil"
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
	Children []*SchemaChild `parser:"@@*"`
}

// GetIncludes returns all include declarations in the schema.
func (s *Schema) GetIncludes() []*Include {
	includes := []*Include{}
	for _, node := range s.Children {
		if node.Kind() == SchemaChildKindInclude {
			includes = append(includes, node.Include)
		}
	}
	return includes
}

// GetComments returns all comments in the schema.
func (s *Schema) GetComments() []*Comment {
	comments := []*Comment{}
	for _, node := range s.Children {
		if node.Kind() == SchemaChildKindComment {
			comments = append(comments, node.Comment)
		}
	}
	return comments
}

// GetDocstrings returns all docstrings in the schema.
func (s *Schema) GetDocstrings() []*Docstring {
	docstrings := []*Docstring{}
	for _, node := range s.Children {
		if node.Kind() == SchemaChildKindDocstring {
			docstrings = append(docstrings, node.Docstring)
		}
	}
	return docstrings
}

// GetTypes returns all custom types in the schema.
func (s *Schema) GetTypes() []*TypeDecl {
	types := []*TypeDecl{}
	for _, node := range s.Children {
		if node.Kind() == SchemaChildKindType {
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
	for _, node := range s.Children {
		if node.Kind() == SchemaChildKindConst {
			consts = append(consts, node.Const)
		}
	}
	return consts
}

// GetEnums returns all enum declarations in the schema.
func (s *Schema) GetEnums() []*EnumDecl {
	enums := []*EnumDecl{}
	for _, node := range s.Children {
		if node.Kind() == SchemaChildKindEnum {
			enums = append(enums, node.Enum)
		}
	}
	return enums
}

// GetPatterns returns all pattern declarations in the schema.
func (s *Schema) GetPatterns() []*PatternDecl {
	patterns := []*PatternDecl{}
	for _, node := range s.Children {
		if node.Kind() == SchemaChildKindPattern {
			patterns = append(patterns, node.Pattern)
		}
	}
	return patterns
}

// GetRPCs returns all RPC blocks in the schema.
func (s *Schema) GetRPCs() []*RPCDecl {
	rpcs := []*RPCDecl{}
	for _, node := range s.Children {
		if node.Kind() == SchemaChildKindRPC {
			rpcs = append(rpcs, node.RPC)
		}
	}
	return rpcs
}

// GetRPCsMap returns a map of RPC names to RPC declarations.
func (s *Schema) GetRPCsMap() map[string]*RPCDecl {
	rpcsMap := make(map[string]*RPCDecl)
	for _, rpc := range s.GetRPCs() {
		rpcsMap[rpc.Name] = rpc
	}
	return rpcsMap
}

// SchemaChildKind represents the kind of a schema child node.
type SchemaChildKind string

const (
	SchemaChildKindInclude   SchemaChildKind = "Include"
	SchemaChildKindComment   SchemaChildKind = "Comment"
	SchemaChildKindDocstring SchemaChildKind = "Docstring"
	SchemaChildKindType      SchemaChildKind = "Type"
	SchemaChildKindConst     SchemaChildKind = "Const"
	SchemaChildKindEnum      SchemaChildKind = "Enum"
	SchemaChildKindPattern   SchemaChildKind = "Pattern"
	SchemaChildKindRPC       SchemaChildKind = "RPC"
)

// SchemaChild represents a child node of the Schema root node.
type SchemaChild struct {
	Positions
	Include   *Include     `parser:"  @@"`
	Comment   *Comment     `parser:"| @@"`
	Type      *TypeDecl    `parser:"| @@"`
	Const     *ConstDecl   `parser:"| @@"`
	Enum      *EnumDecl    `parser:"| @@"`
	Pattern   *PatternDecl `parser:"| @@"`
	RPC       *RPCDecl     `parser:"| @@"`
	Docstring *Docstring   `parser:"| @@"`
}

func (n *SchemaChild) Kind() SchemaChildKind {
	if n.Include != nil {
		return SchemaChildKindInclude
	}
	if n.Comment != nil {
		return SchemaChildKindComment
	}
	if n.Docstring != nil {
		return SchemaChildKindDocstring
	}
	if n.Type != nil {
		return SchemaChildKindType
	}
	if n.Const != nil {
		return SchemaChildKindConst
	}
	if n.Enum != nil {
		return SchemaChildKindEnum
	}
	if n.Pattern != nil {
		return SchemaChildKindPattern
	}
	if n.RPC != nil {
		return SchemaChildKindRPC
	}
	return ""
}

// Include represents an include statement.
type Include struct {
	Positions
	Path QuotedString `parser:"Include @StringLiteral"`
}

// Comment represents both simple and block comments in the schema.
type Comment struct {
	Positions
	Simple *string `parser:"  @Comment"`
	Block  *string `parser:"| @CommentBlock"`
}

// TypeDecl represents a custom type declaration.
type TypeDecl struct {
	Positions
	Docstring  *Docstring       `parser:"(@@ (?! Newline Newline))?"`
	Deprecated *Deprecated      `parser:"@@?"`
	Name       string           `parser:"Type @Ident"`
	Children   []*TypeDeclChild `parser:"LBrace @@* RBrace"`
}

// TypeDeclChild represents a child within a type declaration block.
// Can be a Comment, a Field, or a Spread.
type TypeDeclChild struct {
	Positions
	Comment *Comment `parser:"  @@"`
	Field   *Field   `parser:"| @@"`
	Spread  *Spread  `parser:"| @@"`
}

// GetFlattenedFields returns a recursive flattened list of all fields in the type declaration.
func (t *TypeDecl) GetFlattenedFields() []*Field {
	fields := []*Field{}
	for _, child := range t.Children {
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
	Docstring  *Docstring  `parser:"(@@ (?! Newline Newline))?"`
	Deprecated *Deprecated `parser:"@@?"`
	Name       string      `parser:"Const @Ident"`
	Value      *ConstValue `parser:"Equals @@"`
}

// ConstValue represents the value of a constant.
type ConstValue struct {
	Positions
	Str   *QuotedString `parser:"  @StringLiteral"`
	Float *string       `parser:"| @FloatLiteral"`
	Int   *string       `parser:"| @IntLiteral"`
	True  bool          `parser:"| @True"`
	False bool          `parser:"| @False"`
}

// String returns the string representation of the constant value.
func (cv ConstValue) String() string {
	if cv.Str != nil {
		return `"` + strutil.EscapeQuotes(string(*cv.Str)) + `"`
	}
	if cv.Float != nil {
		return *cv.Float
	}
	if cv.Int != nil {
		return *cv.Int
	}
	if cv.True {
		return "true"
	}
	if cv.False {
		return "false"
	}
	return ""
}

// EnumDecl represents an enum declaration.
type EnumDecl struct {
	Positions
	Docstring  *Docstring    `parser:"(@@ (?! Newline Newline))?"`
	Deprecated *Deprecated   `parser:"@@?"`
	Name       string        `parser:"Enum @Ident"`
	Members    []*EnumMember `parser:"LBrace @@* RBrace"`
}

// EnumMember represents a member of an enum.
type EnumMember struct {
	Positions
	Comment *Comment   `parser:"  @@"`
	Name    string     `parser:"| @Ident"`
	Value   *EnumValue `parser:"  (Equals @@)?"`
}

// EnumValue represents the value of an enum member.
type EnumValue struct {
	Positions
	Str *QuotedString `parser:"  @StringLiteral"`
	Int *string       `parser:"| @IntLiteral"`
}

// PatternDecl represents a pattern declaration.
type PatternDecl struct {
	Positions
	Docstring  *Docstring   `parser:"(@@ (?! Newline Newline))?"`
	Deprecated *Deprecated  `parser:"@@?"`
	Name       string       `parser:"Pattern @Ident"`
	Value      QuotedString `parser:"Equals @StringLiteral"`
}

// RPCDecl represents an RPC service declaration containing procedures and streams.
type RPCDecl struct {
	Positions
	Docstring  *Docstring  `parser:"(@@ (?! Newline Newline))?"`
	Deprecated *Deprecated `parser:"@@?"`
	Name       string      `parser:"Rpc @Ident"`
	Children   []*RPCChild `parser:"LBrace @@* RBrace"`
}

// GetProcs returns all procedures in this RPC block.
func (r *RPCDecl) GetProcs() []*ProcDecl {
	procs := []*ProcDecl{}
	for _, child := range r.Children {
		if child.Proc != nil {
			procs = append(procs, child.Proc)
		}
	}
	return procs
}

// GetStreams returns all streams in this RPC block.
func (r *RPCDecl) GetStreams() []*StreamDecl {
	streams := []*StreamDecl{}
	for _, child := range r.Children {
		if child.Stream != nil {
			streams = append(streams, child.Stream)
		}
	}
	return streams
}

// RPCChild represents a child within an RPC block.
// The order of alternatives is important: Proc/Stream must come before Docstring
// so that ProcDecl/StreamDecl can capture attached docstrings (via their own grammar).
type RPCChild struct {
	Positions
	Comment   *Comment    `parser:"  @@"`
	Proc      *ProcDecl   `parser:"| @@"`
	Stream    *StreamDecl `parser:"| @@"`
	Docstring *Docstring  `parser:"| @@"`
}

// ProcDecl represents a procedure declaration.
type ProcDecl struct {
	Positions
	Docstring  *Docstring               `parser:"(@@ (?! Newline Newline))?"`
	Deprecated *Deprecated              `parser:"@@?"`
	Name       string                   `parser:"Proc @Ident"`
	Children   []*ProcOrStreamDeclChild `parser:"LBrace @@* RBrace"`
}

// StreamDecl represents a stream declaration.
type StreamDecl struct {
	Positions
	Docstring  *Docstring               `parser:"(@@ (?! Newline Newline))?"`
	Deprecated *Deprecated              `parser:"@@?"`
	Name       string                   `parser:"Stream @Ident"`
	Children   []*ProcOrStreamDeclChild `parser:"LBrace @@* RBrace"`
}

// ProcOrStreamDeclChild represents a child node within a ProcDecl or StreamDecl block (Comment, Input, or Output).
type ProcOrStreamDeclChild struct {
	Positions
	Comment *Comment                     `parser:"  @@"`
	Input   *ProcOrStreamDeclChildInput  `parser:"| @@"`
	Output  *ProcOrStreamDeclChildOutput `parser:"| @@"`
}

// ProcOrStreamDeclChildInput represents the Input{...} block within a ProcDecl or StreamDecl.
type ProcOrStreamDeclChildInput struct {
	Positions
	Children []*InputOutputChild `parser:"Input LBrace @@* RBrace"`
}

// GetFlattenedFields returns a recursive flattened list of all fields in the input block.
func (i *ProcOrStreamDeclChildInput) GetFlattenedFields() []*Field {
	fields := []*Field{}
	for _, child := range i.Children {
		if child.Field == nil {
			continue
		}
		fields = append(fields, child.Field.GetFlattenedField()...)
	}
	return fields
}

// ProcOrStreamDeclChildOutput represents the Output{...} block within a ProcDecl or StreamDecl.
type ProcOrStreamDeclChildOutput struct {
	Positions
	Children []*InputOutputChild `parser:"Output LBrace @@* RBrace"`
}

// GetFlattenedFields returns a recursive flattened list of all fields in the output block.
func (o *ProcOrStreamDeclChildOutput) GetFlattenedFields() []*Field {
	fields := []*Field{}
	for _, child := range o.Children {
		if child.Field == nil {
			continue
		}
		fields = append(fields, child.Field.GetFlattenedField()...)
	}
	return fields
}

// InputOutputChild represents a child within an input or output block.
// Can be a Comment, Field, or Spread.
type InputOutputChild struct {
	Positions
	Comment *Comment `parser:"  @@"`
	Field   *Field   `parser:"| @@"`
	Spread  *Spread  `parser:"| @@"`
}

//////////////////
// SHARED TYPES //
//////////////////

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

// Deprecated represents a deprecated declaration.
type Deprecated struct {
	Positions
	Message *QuotedString `parser:"Deprecated (LParen @StringLiteral RParen)?"`
}

// Spread represents a spread operator for type composition (...TypeName).
type Spread struct {
	Positions
	TypeName string `parser:"Spread @Ident"`
}

// Field represents a field definition.
type Field struct {
	Positions
	Docstring *Docstring `parser:"(@@ (?! Newline Newline))?"`
	Name      string     `parser:"@Ident"`
	Optional  bool       `parser:"@Question?"`
	Type      FieldType  `parser:"Colon @@"`
}

// GetFlattenedField returns a recursive flattened list of this field and all its children fields.
func (f *Field) GetFlattenedField() []*Field {
	fields := []*Field{f}

	if f.Type.Base == nil || f.Type.Base.Object == nil {
		return fields
	}

	for _, child := range f.Type.Base.Object.Children {
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

// FieldTypeMap represents a map type: map<ValueType>
type FieldTypeMap struct {
	Positions
	ValueType *FieldType `parser:"Map LessThan @@ GreaterThan"`
}

// FieldTypeObject represents an inline object type definition.
type FieldTypeObject struct {
	Positions
	Children []*TypeDeclChild `parser:"LBrace @@* RBrace"`
}

// AnyLiteral represents any of the built-in literal types.
type AnyLiteral struct {
	Positions
	Str   *QuotedString `parser:"  @StringLiteral"`
	Float *string       `parser:"| @FloatLiteral"`
	Int   *string       `parser:"| @IntLiteral"`
	True  bool          `parser:"| @True"`
	False bool          `parser:"| @False"`
}

// String returns the string representation of the value of the literal.
func (al AnyLiteral) String() string {
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
	return ""
}
