package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/ast"
	"github.com/varavelio/vdl/toolchain/internal/util/testutil"
)

func qptr(s string) *ast.QuotedString {
	q := ast.QuotedString(s)
	return &q
}

func TestParserInclude(t *testing.T) {
	t.Run("Basic include", func(t *testing.T) {
		input := `include "./foo.vdl"`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Include: &ast.Include{Path: "./foo.vdl"}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Multiple includes", func(t *testing.T) {
		input := `
			include "./foo.vdl"
			include "./bar.vdl"
			include "../common/types.vdl"
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{
			{Include: &ast.Include{Path: "./foo.vdl"}},
			{Include: &ast.Include{Path: "./bar.vdl"}},
			{Include: &ast.Include{Path: "../common/types.vdl"}},
		}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

func TestParserConstDecl(t *testing.T) {
	t.Run("String constant", func(t *testing.T) {
		input := `const API_VERSION = "1.0.0"`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Const: &ast.ConstDecl{
			Name:  "API_VERSION",
			Value: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{Str: qptr("1.0.0")}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Integer constant", func(t *testing.T) {
		input := `const MAX_PAGE_SIZE = 100`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Const: &ast.ConstDecl{
			Name:  "MAX_PAGE_SIZE",
			Value: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{Int: new("100")}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Float constant", func(t *testing.T) {
		input := `const DEFAULT_TAX_RATE = 0.21`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Const: &ast.ConstDecl{
			Name:  "DEFAULT_TAX_RATE",
			Value: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{Float: new("0.21")}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Boolean constants", func(t *testing.T) {
		input := `
			const FEATURE_FLAG_ENABLED = true
			const DEBUG_MODE = false
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{
			{Const: &ast.ConstDecl{Name: "FEATURE_FLAG_ENABLED", Value: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{True: true}}}},
			{Const: &ast.ConstDecl{Name: "DEBUG_MODE", Value: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{False: true}}}},
		}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Constant with optional explicit type", func(t *testing.T) {
		input := `const appConfig AppConfigType = { port 8080 }`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Const: &ast.ConstDecl{
			Name:     "appConfig",
			TypeName: new("AppConfigType"),
			Value: &ast.DataLiteral{Object: &ast.DataLiteralObject{Entries: []*ast.DataLiteralObjectEntry{
				{Key: "port", Value: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{Int: new("8080")}}},
			}}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Constant with docstring", func(t *testing.T) {
		input := `
			""" The maximum number of items allowed per request. """
			const MAX_ITEMS = 50
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Const: &ast.ConstDecl{
			Docstring: &ast.Docstring{Value: " The maximum number of items allowed per request. "},
			Name:      "MAX_ITEMS",
			Value:     &ast.DataLiteral{Scalar: &ast.ScalarLiteral{Int: new("50")}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Constant annotations with primitive payload", func(t *testing.T) {
		input := `@deprecated("Use NEW_LIMIT instead") const OLD_LIMIT = 100`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Const: &ast.ConstDecl{
			Annotations: []*ast.Annotation{{
				Name:     "deprecated",
				Argument: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{Str: qptr("Use NEW_LIMIT instead")}},
			}},
			Name:  "OLD_LIMIT",
			Value: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{Int: new("100")}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Constant with object spread and arrays without commas", func(t *testing.T) {
		input := `
			const appConfig = {
				...baseConfig
				port 8080
				targets [
					{ go { output "./gen/go" } }
				]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		require.Len(t, parsed.Declarations, 1)
		decl := parsed.Declarations[0].Const
		require.NotNil(t, decl)
		require.Nil(t, decl.TypeName)
		require.NotNil(t, decl.Value)
		require.NotNil(t, decl.Value.Object)
		require.Len(t, decl.Value.Object.Entries, 3)
		require.Equal(t, "baseConfig", decl.Value.Object.Entries[0].Spread.Ref.Name)
		require.Equal(t, "port", decl.Value.Object.Entries[1].Key)
		require.Equal(t, "8080", *decl.Value.Object.Entries[1].Value.Scalar.Int)
		require.Equal(t, "targets", decl.Value.Object.Entries[2].Key)
		require.NotNil(t, decl.Value.Object.Entries[2].Value.Array)
		require.Len(t, decl.Value.Object.Entries[2].Value.Array.Elements, 1)
	})
}

func TestParserEnumDecl(t *testing.T) {
	t.Run("String enum with implicit values", func(t *testing.T) {
		input := `
			enum OrderStatus {
				Pending
				Processing
				Shipped
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Enum: &ast.EnumDecl{
			Name: "OrderStatus",
			Members: []*ast.EnumMember{
				{Name: "Pending"},
				{Name: "Processing"},
				{Name: "Shipped"},
			},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Enum with explicit values and spread", func(t *testing.T) {
		input := `
			@roleSet
			enum AllRoles {
				SuperAdmin = "super"
				...StandardRoles
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Enum: &ast.EnumDecl{
			Annotations: []*ast.Annotation{{Name: "roleSet"}},
			Name:        "AllRoles",
			Members: []*ast.EnumMember{
				{Name: "SuperAdmin", Value: &ast.EnumValue{Str: qptr("super")}},
				{Spread: &ast.Spread{Ref: &ast.Reference{Name: "StandardRoles"}}},
			},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Integer enum", func(t *testing.T) {
		input := `
			enum Priority {
				// low level
				Low = 1
				High = 10
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Enum: &ast.EnumDecl{
			Name: "Priority",
			Members: []*ast.EnumMember{
				{Name: "Low", Value: &ast.EnumValue{Int: new("1")}},
				{Name: "High", Value: &ast.EnumValue{Int: new("10")}},
			},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

func TestParserEnumMemberAnnotations(t *testing.T) {
	t.Run("Single flag annotation on enum member", func(t *testing.T) {
		input := `
			enum Status {
				@deprecated
				Active
				Inactive
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Enum: &ast.EnumDecl{
			Name: "Status",
			Members: []*ast.EnumMember{
				{Annotations: []*ast.Annotation{{Name: "deprecated"}}, Name: "Active"},
				{Name: "Inactive"},
			},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Annotation with string payload on enum member", func(t *testing.T) {
		input := `
			enum Color {
				@deprecated("Use Crimson instead")
				Red = "red"
				Blue = "blue"
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)
		require.Len(t, members[0].Annotations, 1)
		require.Equal(t, "deprecated", members[0].Annotations[0].Name)
		require.NotNil(t, members[0].Annotations[0].Argument)
		require.Equal(t, "Use Crimson instead", members[0].Annotations[0].Argument.Scalar.Str.String())
		require.Equal(t, "Red", members[0].Name)
		require.Equal(t, "red", members[0].Value.Str.String())
		require.Len(t, members[1].Annotations, 0)
	})

	t.Run("Multiple annotations on enum member", func(t *testing.T) {
		input := `
			enum Permission {
				@deprecated("Use ReadWrite")
				@alias("r")
				Read = "read"
				Write = "write"
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)
		require.Len(t, members[0].Annotations, 2)
		require.Equal(t, "deprecated", members[0].Annotations[0].Name)
		require.Equal(t, "alias", members[0].Annotations[1].Name)
		require.Equal(t, "r", members[0].Annotations[1].Argument.Scalar.Str.String())
		require.Equal(t, "Read", members[0].Name)
	})

	t.Run("Annotation with integer payload on enum member", func(t *testing.T) {
		input := `
			enum Priority {
				@weight(10)
				High = 3
				@weight(1)
				Low = 1
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)
		require.Len(t, members[0].Annotations, 1)
		require.Equal(t, "10", *members[0].Annotations[0].Argument.Scalar.Int)
		require.Equal(t, "High", members[0].Name)
		require.Len(t, members[1].Annotations, 1)
		require.Equal(t, "1", *members[1].Annotations[0].Argument.Scalar.Int)
		require.Equal(t, "Low", members[1].Name)
	})

	t.Run("Annotation with object payload on enum member", func(t *testing.T) {
		input := `
			enum Role {
				@meta({
					description "Full access"
					level 100
				})
				Admin = "admin"
				User = "user"
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)
		require.Len(t, members[0].Annotations, 1)
		ann := members[0].Annotations[0]
		require.Equal(t, "meta", ann.Name)
		require.NotNil(t, ann.Argument.Object)
		require.Len(t, ann.Argument.Object.Entries, 2)
		require.Equal(t, "description", ann.Argument.Object.Entries[0].Key)
		require.Equal(t, "Full access", ann.Argument.Object.Entries[0].Value.Scalar.Str.String())
		require.Equal(t, "level", ann.Argument.Object.Entries[1].Key)
		require.Equal(t, "100", *ann.Argument.Object.Entries[1].Value.Scalar.Int)
	})

	t.Run("Annotation with array payload on enum member", func(t *testing.T) {
		input := `
			enum Feature {
				@tags(["core" "stable"])
				Auth
				@tags(["beta"])
				Experimental
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)
		require.Len(t, members[0].Annotations, 1)
		require.NotNil(t, members[0].Annotations[0].Argument.Array)
		require.Len(t, members[0].Annotations[0].Argument.Array.Elements, 2)
		require.Equal(t, "core", members[0].Annotations[0].Argument.Array.Elements[0].Scalar.Str.String())
		require.Equal(t, "stable", members[0].Annotations[0].Argument.Array.Elements[1].Scalar.Str.String())
		require.Len(t, members[1].Annotations, 1)
		require.Len(t, members[1].Annotations[0].Argument.Array.Elements, 1)
	})

	t.Run("Annotation with boolean payload on enum member", func(t *testing.T) {
		input := `
			enum Mode {
				@hidden(true)
				Debug
				@hidden(false)
				Release
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)
		require.True(t, members[0].Annotations[0].Argument.Scalar.True)
		require.True(t, members[1].Annotations[0].Argument.Scalar.False)
	})

	t.Run("Docstring on enum member", func(t *testing.T) {
		input := `
			enum Status {
				""" Active means the user is live """
				Active
				""" Banned means the user is blocked """
				Banned
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)
		require.NotNil(t, members[0].Docstring)
		require.Equal(t, " Active means the user is live ", members[0].Docstring.Value.String())
		require.Equal(t, "Active", members[0].Name)
		require.NotNil(t, members[1].Docstring)
		require.Equal(t, " Banned means the user is blocked ", members[1].Docstring.Value.String())
		require.Equal(t, "Banned", members[1].Name)
	})

	t.Run("Docstring with blank line inside enum fails", func(t *testing.T) {
		// Unlike type bodies which support standalone docstrings as TypeMember,
		// enum bodies do not have a standalone docstring alternative.
		// A docstring separated by a blank line from the next member is invalid.
		input := `
			enum Status {
				""" Section header """

				Active
			}
		`
		_, err := ParserInstance.ParseString("schema.vdl", input)
		require.Error(t, err)
	})

	t.Run("Docstring and annotation on enum member", func(t *testing.T) {
		input := `
			enum Status {
				""" The active state """
				@live
				Active
				Inactive
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)
		require.NotNil(t, members[0].Docstring)
		require.Equal(t, " The active state ", members[0].Docstring.Value.String())
		require.Len(t, members[0].Annotations, 1)
		require.Equal(t, "live", members[0].Annotations[0].Name)
		require.Equal(t, "Active", members[0].Name)
		require.Nil(t, members[1].Docstring)
		require.Len(t, members[1].Annotations, 0)
		require.Equal(t, "Inactive", members[1].Name)
	})

	t.Run("Docstring and annotation on enum member with value", func(t *testing.T) {
		input := `
			enum HttpCode {
				""" Success """
				@standard
				OK = 200
				""" Not found """
				@standard
				NotFound = 404
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)

		require.NotNil(t, members[0].Docstring)
		require.Equal(t, " Success ", members[0].Docstring.Value.String())
		require.Len(t, members[0].Annotations, 1)
		require.Equal(t, "standard", members[0].Annotations[0].Name)
		require.Equal(t, "OK", members[0].Name)
		require.Equal(t, "200", *members[0].Value.Int)

		require.NotNil(t, members[1].Docstring)
		require.Equal(t, " Not found ", members[1].Docstring.Value.String())
		require.Len(t, members[1].Annotations, 1)
		require.Equal(t, "NotFound", members[1].Name)
		require.Equal(t, "404", *members[1].Value.Int)
	})

	t.Run("Enum with both type-level and member-level annotations", func(t *testing.T) {
		input := `
			""" Role definitions """
			@rbac
			enum Role {
				@meta({ level 100 })
				Admin = "admin"
				@meta({ level 10 })
				User = "user"
				Guest = "guest"
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		enumDecl := parsed.Declarations[0].Enum
		require.NotNil(t, enumDecl.Docstring)
		require.Equal(t, " Role definitions ", enumDecl.Docstring.Value.String())
		require.Len(t, enumDecl.Annotations, 1)
		require.Equal(t, "rbac", enumDecl.Annotations[0].Name)

		members := enumDecl.Members
		require.Len(t, members, 3)
		require.Len(t, members[0].Annotations, 1)
		require.Equal(t, "meta", members[0].Annotations[0].Name)
		require.Equal(t, "Admin", members[0].Name)
		require.Len(t, members[1].Annotations, 1)
		require.Equal(t, "User", members[1].Name)
		require.Len(t, members[2].Annotations, 0)
		require.Equal(t, "Guest", members[2].Name)
	})

	t.Run("Annotation on spread in enum", func(t *testing.T) {
		input := `
			enum AllRoles {
				@deprecated
				SuperAdmin = "super"
				...StandardRoles
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)
		require.Len(t, members[0].Annotations, 1)
		require.Equal(t, "deprecated", members[0].Annotations[0].Name)
		require.Equal(t, "SuperAdmin", members[0].Name)
		require.NotNil(t, members[1].Spread)
		require.Equal(t, "StandardRoles", members[1].Spread.Ref.Name)
	})

	t.Run("Every member annotated", func(t *testing.T) {
		input := `
			enum Direction {
				@axis("x")
				Left
				@axis("x")
				Right
				@axis("y")
				Up
				@axis("y")
				Down
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 4)
		for _, m := range members {
			require.Len(t, m.Annotations, 1)
			require.Equal(t, "axis", m.Annotations[0].Name)
		}
		require.Equal(t, "x", members[0].Annotations[0].Argument.Scalar.Str.String())
		require.Equal(t, "x", members[1].Annotations[0].Argument.Scalar.Str.String())
		require.Equal(t, "y", members[2].Annotations[0].Argument.Scalar.Str.String())
		require.Equal(t, "y", members[3].Annotations[0].Argument.Scalar.Str.String())
	})

	t.Run("Annotation with reference payload on enum member", func(t *testing.T) {
		input := `
			enum Theme {
				@fallback(DEFAULT_THEME)
				Dark
				Light
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 2)
		require.Len(t, members[0].Annotations, 1)
		ann := members[0].Annotations[0]
		require.Equal(t, "fallback", ann.Name)
		require.NotNil(t, ann.Argument.Scalar.Ref)
		require.Equal(t, "DEFAULT_THEME", ann.Argument.Scalar.Ref.Name)
		require.Nil(t, ann.Argument.Scalar.Ref.Member)
	})
}

func TestParserTypeDecl(t *testing.T) {
	t.Run("Minimum type declaration", func(t *testing.T) {
		input := `
			type MyType {
				field string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Type: &ast.TypeDecl{
			Name: "MyType",
			Members: []*ast.TypeMember{{
				Field: &ast.Field{Name: "field", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: new("string")}}},
			}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with docstring and annotation", func(t *testing.T) {
		input := `
			""" My type description """
			@entity
			type MyType {
				field string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Type: &ast.TypeDecl{
			Docstring:   &ast.Docstring{Value: " My type description "},
			Annotations: []*ast.Annotation{{Name: "entity"}},
			Name:        "MyType",
			Members: []*ast.TypeMember{{
				Field: &ast.Field{Name: "field", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: new("string")}}},
			}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with all primitive fields and optional field", func(t *testing.T) {
		input := `
			type MyType {
				f1 string
				f2 int
				f3 float
				f4 bool
				f5 datetime
				optional? string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		require.Len(t, parsed.Declarations, 1)
		typeDecl := parsed.Declarations[0].Type
		require.Equal(t, "MyType", typeDecl.Name)
		require.Len(t, typeDecl.Members, 6)
		require.True(t, typeDecl.Members[5].Field.Optional)
	})

	t.Run("Type with arrays and multidimensional arrays", func(t *testing.T) {
		input := `
			type MyType {
				tags string[]
				scores int[][]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Type: &ast.TypeDecl{
			Name: "MyType",
			Members: []*ast.TypeMember{
				{Field: &ast.Field{Name: "tags", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: new("string")}, Dimensions: 1}}},
				{Field: &ast.Field{Name: "scores", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: new("int")}, Dimensions: 2}}},
			},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with nested inline object", func(t *testing.T) {
		input := `
			type MyType {
				location {
					lat float
					lng float
				}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		require.Len(t, parsed.Declarations, 1)
		typeDecl := parsed.Declarations[0].Type
		require.Len(t, typeDecl.Members, 1)
		location := typeDecl.Members[0].Field
		require.Equal(t, "location", location.Name)
		require.NotNil(t, location.Type.Base.Object)
		require.Len(t, location.Type.Base.Object.Members, 2)
	})

	t.Run("Type with map fields", func(t *testing.T) {
		input := `
			type MyType {
				counts map[int]
				users map[User]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Type: &ast.TypeDecl{
			Name: "MyType",
			Members: []*ast.TypeMember{
				{Field: &ast.Field{Name: "counts", Type: ast.FieldType{Base: &ast.FieldTypeBase{Map: &ast.FieldTypeMap{ValueType: &ast.FieldType{Base: &ast.FieldTypeBase{Named: new("int")}}}}}}},
				{Field: &ast.Field{Name: "users", Type: ast.FieldType{Base: &ast.FieldTypeBase{Map: &ast.FieldTypeMap{ValueType: &ast.FieldType{Base: &ast.FieldTypeBase{Named: new("User")}}}}}}},
			},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Type with spread operator and field annotation", func(t *testing.T) {
		input := `
			type Article {
				...AuditMetadata
				@title
				heading string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Type: &ast.TypeDecl{
			Name: "Article",
			Members: []*ast.TypeMember{
				{Spread: &ast.Spread{Ref: &ast.Reference{Name: "AuditMetadata"}}},
				{Field: &ast.Field{Annotations: []*ast.Annotation{{Name: "title"}}, Name: "heading", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: new("string")}}}},
			},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Nested type model for rpc-style shapes", func(t *testing.T) {
		input := `
			@rpc
			type Chat {
				@proc
				SendMessage {
					input {
						chatId string
						message string
					}
					output {
						messageId string
					}
				}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		require.Len(t, parsed.Declarations, 1)
		rpcType := parsed.Declarations[0].Type
		require.Equal(t, "Chat", rpcType.Name)
		require.Len(t, rpcType.Annotations, 1)
		require.Equal(t, "rpc", rpcType.Annotations[0].Name)
		require.Len(t, rpcType.Members, 1)
		sendMessage := rpcType.Members[0].Field
		require.Equal(t, "SendMessage", sendMessage.Name)
		require.Len(t, sendMessage.Annotations, 1)
		require.Equal(t, "proc", sendMessage.Annotations[0].Name)
		require.NotNil(t, sendMessage.Type.Base.Object)
		require.Len(t, sendMessage.Type.Base.Object.Members, 2)
	})
}

func TestParserDocstrings(t *testing.T) {
	t.Run("Standalone docstring", func(t *testing.T) {
		input := `""" This is a standalone docstring. """`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Docstring: &ast.Docstring{Value: " This is a standalone docstring. "}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Multiple standalone docstrings", func(t *testing.T) {
		input := `
			""" First docstring """
			""" Second docstring """
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{
			{Docstring: &ast.Docstring{Value: " First docstring "}},
			{Docstring: &ast.Docstring{Value: " Second docstring "}},
		}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Docstring followed by blank line becomes standalone", func(t *testing.T) {
		input := `
			""" This is standalone """

			type MyType {
				field string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{
			{Docstring: &ast.Docstring{Value: " This is standalone "}},
			{Type: &ast.TypeDecl{Name: "MyType", Members: []*ast.TypeMember{{Field: &ast.Field{Name: "field", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: new("string")}}}}}}},
		}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})
}

func TestParserEdgeCases(t *testing.T) {
	t.Run("Empty schema", func(t *testing.T) {
		parsed, err := ParserInstance.ParseString("schema.vdl", ``)
		require.NoError(t, err)
		testutil.ASTEqualNoPos(t, &ast.Schema{Declarations: nil}, parsed)
	})

	t.Run("Whitespace only", func(t *testing.T) {
		parsed, err := ParserInstance.ParseString("schema.vdl", "   \n\t\n")
		require.NoError(t, err)
		testutil.ASTEqualNoPos(t, &ast.Schema{Declarations: nil}, parsed)
	})

	t.Run("Empty type and enum", func(t *testing.T) {
		input := `
			type EmptyType {}
			enum EmptyEnum {}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{
			{Type: &ast.TypeDecl{Name: "EmptyType", Members: nil}},
			{Enum: &ast.EnumDecl{Name: "EmptyEnum", Members: nil}},
		}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Deeply nested inline objects", func(t *testing.T) {
		input := `
			type Deep {
				level1 {
					level2 {
						value string
					}
				}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		require.Len(t, parsed.Declarations, 1)
		require.Equal(t, "Deep", parsed.Declarations[0].Type.Name)
		require.NotNil(t, parsed.Declarations[0].Type.Members[0].Field.Type.Base.Object)
	})

	t.Run("Array of inline objects", func(t *testing.T) {
		input := `
			type MyType {
				items {
					name string
				}[]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		require.Len(t, parsed.Declarations, 1)
		items := parsed.Declarations[0].Type.Members[0].Field
		require.Equal(t, ast.ArrayDimensions(1), items.Type.Dimensions)
		require.NotNil(t, items.Type.Base.Object)
	})

	t.Run("Map of arrays", func(t *testing.T) {
		input := `
			type MyType {
				data map[string[]]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		require.Len(t, parsed.Declarations, 1)
		field := parsed.Declarations[0].Type.Members[0].Field
		require.NotNil(t, field.Type.Base.Map)
		require.Equal(t, ast.ArrayDimensions(1), field.Type.Base.Map.ValueType.Dimensions)
	})

	t.Run("Custom type reference", func(t *testing.T) {
		input := `
			type Profile {
				user User
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)
		require.Equal(t, "User", *parsed.Declarations[0].Type.Members[0].Field.Type.Base.Named)
	})
}

func TestParserMultiDimensionalArrays(t *testing.T) {
	t.Run("2D and 3D arrays", func(t *testing.T) {
		input := `
			type Container {
				single int
				oneDim int[]
				twoDim int[][]
				threeDim int[][][]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		require.Equal(t, ast.ArrayDimensions(0), typeDecl.Members[0].Field.Type.Dimensions)
		require.Equal(t, ast.ArrayDimensions(1), typeDecl.Members[1].Field.Type.Dimensions)
		require.Equal(t, ast.ArrayDimensions(2), typeDecl.Members[2].Field.Type.Dimensions)
		require.Equal(t, ast.ArrayDimensions(3), typeDecl.Members[3].Field.Type.Dimensions)
	})

	t.Run("Map with multi-dimensional array value", func(t *testing.T) {
		input := `
			type Cache {
				matrices map[int[][]][][][]
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		field := parsed.Declarations[0].Type.Members[0].Field
		require.Equal(t, ast.ArrayDimensions(3), field.Type.Dimensions)
		require.NotNil(t, field.Type.Base.Map)
		require.Equal(t, ast.ArrayDimensions(2), field.Type.Base.Map.ValueType.Dimensions)
	})
}

func TestParserKeywordsAsFieldNames(t *testing.T) {
	t.Run("Current and old keywords as field names in type", func(t *testing.T) {
		input := `
			type KeywordFields {
				type string
				enum string
				const string
				include string
				map string
				string string
				int string
				float string
				bool string
				datetime string
				true string
				false string
				rpc string
				proc string
				stream string
				input string
				output string
				pattern string
				deprecated string
			}
		`

		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)
		require.Len(t, parsed.Declarations, 1)
		require.Len(t, parsed.Declarations[0].Type.Members, 19)
	})

	t.Run("Keywords as optional field names", func(t *testing.T) {
		input := `
			type OptionalKeywords {
				input? string
				output? int
				type? bool
			}
		`

		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)
		typeDecl := parsed.Declarations[0].Type
		require.True(t, typeDecl.Members[0].Field.Optional)
		require.True(t, typeDecl.Members[1].Field.Optional)
		require.True(t, typeDecl.Members[2].Field.Optional)
	})
}

func TestParserAnnotations(t *testing.T) {
	t.Run("Flags and primitive payloads", func(t *testing.T) {
		input := `
			@entity
			@deprecated("Use UserV2")
			type User {
				@id
				id string
			}
		`

		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		require.Len(t, typeDecl.Annotations, 2)
		require.Equal(t, "entity", typeDecl.Annotations[0].Name)
		require.Equal(t, "deprecated", typeDecl.Annotations[1].Name)
		require.NotNil(t, typeDecl.Annotations[1].Argument)
		require.Equal(t, "Use UserV2", typeDecl.Annotations[1].Argument.Scalar.Str.String())
		require.Len(t, typeDecl.Members[0].Field.Annotations, 1)
		require.Equal(t, "id", typeDecl.Members[0].Field.Annotations[0].Name)
	})

	t.Run("Complex annotation payload", func(t *testing.T) {
		input := `
			@auth({
				roles ["admin" "user"]
				rules {
					strict true
				}
			})
			type Secured {
				id string
			}
		`

		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		annotation := parsed.Declarations[0].Type.Annotations[0]
		require.Equal(t, "auth", annotation.Name)
		require.NotNil(t, annotation.Argument)
		require.NotNil(t, annotation.Argument.Object)
		require.Len(t, annotation.Argument.Object.Entries, 2)
		require.Equal(t, "roles", annotation.Argument.Object.Entries[0].Key)
		require.NotNil(t, annotation.Argument.Object.Entries[0].Value.Array)
		require.Equal(t, "rules", annotation.Argument.Object.Entries[1].Key)
		require.NotNil(t, annotation.Argument.Object.Entries[1].Value.Object)
	})

	t.Run("Multiple annotations in same line", func(t *testing.T) {
		input := `@a @b("x") type T { f string }`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		require.Len(t, typeDecl.Annotations, 2)
		require.Equal(t, "a", typeDecl.Annotations[0].Name)
		require.Equal(t, "b", typeDecl.Annotations[1].Name)
		require.Equal(t, "x", typeDecl.Annotations[1].Argument.Scalar.Str.String())
	})

	t.Run("Field annotation with complex payload", func(t *testing.T) {
		input := `
			type User {
				@db({
					index true
					tags ["a" "b"]
				})
				email string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		field := parsed.Declarations[0].Type.Members[0].Field
		require.Len(t, field.Annotations, 1)
		ann := field.Annotations[0]
		require.Equal(t, "db", ann.Name)
		require.NotNil(t, ann.Argument)
		require.NotNil(t, ann.Argument.Object)
		require.Len(t, ann.Argument.Object.Entries, 2)
	})
}

func TestParserConstWithReferences(t *testing.T) {
	t.Run("const with enum member reference value", func(t *testing.T) {
		input := `const DEFAULT_COLOR = Color.Red`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Const: &ast.ConstDecl{
			Name: "DEFAULT_COLOR",
			Value: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{Ref: &ast.Reference{
				Name:   "Color",
				Member: new("Red"),
			}}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("const with const reference value", func(t *testing.T) {
		input := `const MY_CONFIG = BASE_CONFIG`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Const: &ast.ConstDecl{
			Name: "MY_CONFIG",
			Value: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{Ref: &ast.Reference{
				Name: "BASE_CONFIG",
			}}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("const with typed enum member reference value", func(t *testing.T) {
		input := `const DEFAULT_STATUS StatusEnum = Status.Active`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Const: &ast.ConstDecl{
			Name:     "DEFAULT_STATUS",
			TypeName: new("StatusEnum"),
			Value: &ast.DataLiteral{Scalar: &ast.ScalarLiteral{Ref: &ast.Reference{
				Name:   "Status",
				Member: new("Active"),
			}}},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("const object with enum member reference in value", func(t *testing.T) {
		input := `
			const appConfig = {
				color Color.Red
				status Status.Active
				name "test"
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		require.Len(t, parsed.Declarations, 1)
		decl := parsed.Declarations[0].Const
		require.NotNil(t, decl.Value.Object)
		entries := decl.Value.Object.Entries
		require.Len(t, entries, 3)

		// First entry: color Color.Red
		require.Equal(t, "color", entries[0].Key)
		require.NotNil(t, entries[0].Value.Scalar.Ref)
		require.Equal(t, "Color", entries[0].Value.Scalar.Ref.Name)
		require.Equal(t, "Red", *entries[0].Value.Scalar.Ref.Member)

		// Second entry: status Status.Active
		require.Equal(t, "status", entries[1].Key)
		require.NotNil(t, entries[1].Value.Scalar.Ref)
		require.Equal(t, "Status", entries[1].Value.Scalar.Ref.Name)
		require.Equal(t, "Active", *entries[1].Value.Scalar.Ref.Member)

		// Third entry: name "test"
		require.Equal(t, "name", entries[2].Key)
		require.NotNil(t, entries[2].Value.Scalar.Str)
		require.Equal(t, "test", entries[2].Value.Scalar.Str.String())
	})

	t.Run("const object with const reference in value", func(t *testing.T) {
		input := `
			const derived = {
				base BASE_CONFIG
				extra "stuff"
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		entries := parsed.Declarations[0].Const.Value.Object.Entries
		require.Len(t, entries, 2)
		require.NotNil(t, entries[0].Value.Scalar.Ref)
		require.Equal(t, "BASE_CONFIG", entries[0].Value.Scalar.Ref.Name)
		require.Nil(t, entries[0].Value.Scalar.Ref.Member)
	})

	t.Run("const array with mixed references and literals", func(t *testing.T) {
		input := `
			const items = [
				Color.Red
				"hello"
				42
				BASE
			]
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		arr := parsed.Declarations[0].Const.Value.Array
		require.NotNil(t, arr)
		require.Len(t, arr.Elements, 4)

		// Color.Red
		require.NotNil(t, arr.Elements[0].Scalar.Ref)
		require.Equal(t, "Color", arr.Elements[0].Scalar.Ref.Name)
		require.Equal(t, "Red", *arr.Elements[0].Scalar.Ref.Member)

		// "hello"
		require.NotNil(t, arr.Elements[1].Scalar.Str)

		// 42
		require.NotNil(t, arr.Elements[2].Scalar.Int)

		// BASE (const ref)
		require.NotNil(t, arr.Elements[3].Scalar.Ref)
		require.Equal(t, "BASE", arr.Elements[3].Scalar.Ref.Name)
		require.Nil(t, arr.Elements[3].Scalar.Ref.Member)
	})
}

func TestParserAnnotationPayloadVariants(t *testing.T) {
	t.Run("annotation with array payload", func(t *testing.T) {
		input := `
			@tags(["admin" "user" "guest"])
			type User {
				id string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		ann := parsed.Declarations[0].Type.Annotations[0]
		require.Equal(t, "tags", ann.Name)
		require.NotNil(t, ann.Argument.Array)
		require.Len(t, ann.Argument.Array.Elements, 3)
		require.Equal(t, "admin", ann.Argument.Array.Elements[0].Scalar.Str.String())
		require.Equal(t, "user", ann.Argument.Array.Elements[1].Scalar.Str.String())
		require.Equal(t, "guest", ann.Argument.Array.Elements[2].Scalar.Str.String())
	})

	t.Run("annotation with enum member reference payload", func(t *testing.T) {
		input := `
			@default(Color.Red)
			type Themed {
				color string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		ann := parsed.Declarations[0].Type.Annotations[0]
		require.Equal(t, "default", ann.Name)
		require.NotNil(t, ann.Argument.Scalar)
		require.NotNil(t, ann.Argument.Scalar.Ref)
		require.Equal(t, "Color", ann.Argument.Scalar.Ref.Name)
		require.Equal(t, "Red", *ann.Argument.Scalar.Ref.Member)
	})

	t.Run("annotation with const reference payload", func(t *testing.T) {
		input := `
			@fallback(DEFAULT_CONFIG)
			type Settings {
				value string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		ann := parsed.Declarations[0].Type.Annotations[0]
		require.Equal(t, "fallback", ann.Name)
		require.NotNil(t, ann.Argument.Scalar.Ref)
		require.Equal(t, "DEFAULT_CONFIG", ann.Argument.Scalar.Ref.Name)
		require.Nil(t, ann.Argument.Scalar.Ref.Member)
	})

	t.Run("annotation with integer payload", func(t *testing.T) {
		input := `
			@maxItems(100)
			type List {
				items string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		ann := parsed.Declarations[0].Type.Annotations[0]
		require.Equal(t, "maxItems", ann.Name)
		require.NotNil(t, ann.Argument.Scalar.Int)
		require.Equal(t, "100", *ann.Argument.Scalar.Int)
	})

	t.Run("annotation with boolean payload", func(t *testing.T) {
		input := `
			@strict(true)
			type Validated {
				value string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		ann := parsed.Declarations[0].Type.Annotations[0]
		require.Equal(t, "strict", ann.Name)
		require.NotNil(t, ann.Argument.Scalar)
		require.True(t, ann.Argument.Scalar.True)
	})

	t.Run("field annotation with reference payload", func(t *testing.T) {
		input := `
			type User {
				@default(Status.Active)
				status string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		field := parsed.Declarations[0].Type.Members[0].Field
		require.Len(t, field.Annotations, 1)
		ann := field.Annotations[0]
		require.Equal(t, "default", ann.Name)
		require.NotNil(t, ann.Argument.Scalar.Ref)
		require.Equal(t, "Status", ann.Argument.Scalar.Ref.Name)
		require.Equal(t, "Active", *ann.Argument.Scalar.Ref.Member)
	})

	t.Run("annotation with nested object containing references", func(t *testing.T) {
		input := `
			@config({
				defaultColor Color.Red
				fallback BASE_CONFIG
				enabled true
			})
			type App {
				id string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		ann := parsed.Declarations[0].Type.Annotations[0]
		require.Equal(t, "config", ann.Name)
		require.NotNil(t, ann.Argument.Object)
		entries := ann.Argument.Object.Entries
		require.Len(t, entries, 3)

		// defaultColor Color.Red
		require.NotNil(t, entries[0].Value.Scalar.Ref)
		require.Equal(t, "Color", entries[0].Value.Scalar.Ref.Name)
		require.Equal(t, "Red", *entries[0].Value.Scalar.Ref.Member)

		// fallback BASE_CONFIG
		require.NotNil(t, entries[1].Value.Scalar.Ref)
		require.Equal(t, "BASE_CONFIG", entries[1].Value.Scalar.Ref.Name)
		require.Nil(t, entries[1].Value.Scalar.Ref.Member)

		// enabled true
		require.True(t, entries[2].Value.Scalar.True)
	})
}

func TestParserDocstringInsideTypeBody(t *testing.T) {
	t.Run("docstring without blank line attaches to next field", func(t *testing.T) {
		input := `
			type User {
				""" This docstring attaches to id """
				id string
				name string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		// Without a blank line, the docstring attaches to the field (same as top-level behavior)
		require.Len(t, typeDecl.Members, 2)
		require.NotNil(t, typeDecl.Members[0].Field)
		require.NotNil(t, typeDecl.Members[0].Field.Docstring)
		require.Equal(t, " This docstring attaches to id ", typeDecl.Members[0].Field.Docstring.Value.String())
		require.Equal(t, "id", typeDecl.Members[0].Field.Name)
		require.NotNil(t, typeDecl.Members[1].Field)
		require.Equal(t, "name", typeDecl.Members[1].Field.Name)
	})

	t.Run("docstring attached to field inside type body", func(t *testing.T) {
		input := `
			type User {
				""" The user ID """
				id string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		require.Len(t, typeDecl.Members, 1)
		field := typeDecl.Members[0].Field
		require.NotNil(t, field)
		require.NotNil(t, field.Docstring)
		require.Equal(t, " The user ID ", field.Docstring.Value.String())
		require.Equal(t, "id", field.Name)
	})

	t.Run("docstring with blank line becomes standalone in type body", func(t *testing.T) {
		input := `
			type User {
				""" Section header """

				id string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		require.Len(t, typeDecl.Members, 2)
		require.NotNil(t, typeDecl.Members[0].Docstring)
		require.Equal(t, " Section header ", typeDecl.Members[0].Docstring.Value.String())
		require.NotNil(t, typeDecl.Members[1].Field)
		require.Equal(t, "id", typeDecl.Members[1].Field.Name)
	})

	t.Run("multiple docstrings in type body", func(t *testing.T) {
		input := `
			type User {
				""" Identity fields """

				id string
				name string

				""" Contact info """

				email string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		require.Len(t, typeDecl.Members, 5)
		require.NotNil(t, typeDecl.Members[0].Docstring)
		require.Equal(t, " Identity fields ", typeDecl.Members[0].Docstring.Value.String())
		require.NotNil(t, typeDecl.Members[1].Field)
		require.Equal(t, "id", typeDecl.Members[1].Field.Name)
		require.NotNil(t, typeDecl.Members[2].Field)
		require.Equal(t, "name", typeDecl.Members[2].Field.Name)
		require.NotNil(t, typeDecl.Members[3].Docstring)
		require.Equal(t, " Contact info ", typeDecl.Members[3].Docstring.Value.String())
		require.NotNil(t, typeDecl.Members[4].Field)
		require.Equal(t, "email", typeDecl.Members[4].Field.Name)
	})

	t.Run("docstring attached to annotated field in type body", func(t *testing.T) {
		input := `
			type User {
				""" The email address """
				@unique
				email string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		require.Len(t, typeDecl.Members, 1)
		field := typeDecl.Members[0].Field
		require.NotNil(t, field.Docstring)
		require.Equal(t, " The email address ", field.Docstring.Value.String())
		require.Len(t, field.Annotations, 1)
		require.Equal(t, "unique", field.Annotations[0].Name)
		require.Equal(t, "email", field.Name)
	})
}

func TestParserInvalidSyntax(t *testing.T) {
	t.Run("Field with colon fails", func(t *testing.T) {
		input := `type Bad { name: string }`
		_, err := ParserInstance.ParseString("schema.vdl", input)
		require.Error(t, err)
	})

	t.Run("Top-level legacy rpc fails", func(t *testing.T) {
		input := `rpc Service {}`
		_, err := ParserInstance.ParseString("schema.vdl", input)
		require.Error(t, err)
	})

	t.Run("Array with commas fails", func(t *testing.T) {
		input := `const bad = [1, 2]`
		_, err := ParserInstance.ParseString("schema.vdl", input)
		require.Error(t, err)
	})
}

func TestParserLargeSchema(t *testing.T) {
	input := `
		include "./common.vdl"

		""" service description """
		@rpc
		type Chat {
			@proc
			SendMessage {
				input {
					chatId string
					message string
				}
				output {
					messageId string
				}
			}
			...BaseChat
		}

		@pattern
		const UserTopic = "events.users.{userId}"

		const compilerConfig = {
			version 1
			targets [
				{ go { output "./gen/go" } }
				{ ts { output "./gen/ts" } }
			]
		}

		enum Roles {
			Admin = "admin"
			User = "user"
			...DefaultRoles
		}
	`

	parsed, err := ParserInstance.ParseString("schema.vdl", input)
	require.NoError(t, err)
	require.Len(t, parsed.Declarations, 5)

	require.NotNil(t, parsed.Declarations[0].Include)
	require.NotNil(t, parsed.Declarations[1].Type)
	require.NotNil(t, parsed.Declarations[2].Const)
	require.NotNil(t, parsed.Declarations[3].Const)
	require.NotNil(t, parsed.Declarations[4].Enum)

	chat := parsed.Declarations[1].Type
	require.Equal(t, "Chat", chat.Name)
	require.Len(t, chat.Annotations, 1)
	require.Equal(t, "rpc", chat.Annotations[0].Name)
	require.Len(t, chat.Members, 2)
	require.NotNil(t, chat.Members[1].Spread)

	cfg := parsed.Declarations[3].Const
	require.Equal(t, "compilerConfig", cfg.Name)
	require.NotNil(t, cfg.Value.Object)
	require.Len(t, cfg.Value.Object.Entries, 2)

	roles := parsed.Declarations[4].Enum
	require.Equal(t, "Roles", roles.Name)
	require.Len(t, roles.Members, 3)
	require.NotNil(t, roles.Members[2].Spread)
}

func TestParserDocstringAttachmentWithAnnotations(t *testing.T) {
	t.Run("docstring attaches across annotation lines", func(t *testing.T) {
		input := `
			""" type docs """
			@entity
			type User {
				id string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		require.NotNil(t, typeDecl.Docstring)
		require.Equal(t, " type docs ", typeDecl.Docstring.Value.String())
		require.Len(t, typeDecl.Annotations, 1)
	})

	t.Run("blank line keeps docstring standalone", func(t *testing.T) {
		input := `
			""" docs """

			@entity
			type User {
				id string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)
		require.NotNil(t, parsed.Declarations[0].Docstring)
		require.NotNil(t, parsed.Declarations[1].Type)
	})
}

func TestParserNamespacedSpread(t *testing.T) {
	t.Run("Namespaced spread in type body", func(t *testing.T) {
		input := `
			type User {
				...auth.BaseUser
				name string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Type: &ast.TypeDecl{
			Name: "User",
			Members: []*ast.TypeMember{
				{Spread: &ast.Spread{Ref: &ast.Reference{Name: "auth", Member: new("BaseUser")}}},
				{Field: &ast.Field{Name: "name", Type: ast.FieldType{Base: &ast.FieldTypeBase{Named: new("string")}}}},
			},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Namespaced spread in enum body", func(t *testing.T) {
		input := `
			enum AllRoles {
				Admin = "admin"
				...auth.StandardRoles
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		expected := &ast.Schema{Declarations: []*ast.TopLevelDecl{{Enum: &ast.EnumDecl{
			Name: "AllRoles",
			Members: []*ast.EnumMember{
				{Name: "Admin", Value: &ast.EnumValue{Str: qptr("admin")}},
				{Spread: &ast.Spread{Ref: &ast.Reference{Name: "auth", Member: new("StandardRoles")}}},
			},
		}}}}
		testutil.ASTEqualNoPos(t, expected, parsed)
	})

	t.Run("Namespaced spread in const object literal", func(t *testing.T) {
		input := `
			const config = {
				...shared.baseConfig
				port 8080
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		decl := parsed.Declarations[0].Const
		require.NotNil(t, decl.Value.Object)
		entries := decl.Value.Object.Entries
		require.Len(t, entries, 2)
		require.NotNil(t, entries[0].Spread)
		require.Equal(t, "shared", entries[0].Spread.Ref.Name)
		require.Equal(t, "baseConfig", *entries[0].Spread.Ref.Member)
		require.Equal(t, "port", entries[1].Key)
	})

	t.Run("Mixed local and namespaced spreads in type body", func(t *testing.T) {
		input := `
			type FullUser {
				...BaseUser
				...auth.Credentials
				...billing.PaymentInfo
				name string
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		require.Equal(t, "FullUser", typeDecl.Name)
		require.Len(t, typeDecl.Members, 4)

		// Local spread
		require.NotNil(t, typeDecl.Members[0].Spread)
		require.Equal(t, "BaseUser", typeDecl.Members[0].Spread.Ref.Name)
		require.Nil(t, typeDecl.Members[0].Spread.Ref.Member)

		// Namespaced spread 1
		require.NotNil(t, typeDecl.Members[1].Spread)
		require.Equal(t, "auth", typeDecl.Members[1].Spread.Ref.Name)
		require.Equal(t, "Credentials", *typeDecl.Members[1].Spread.Ref.Member)

		// Namespaced spread 2
		require.NotNil(t, typeDecl.Members[2].Spread)
		require.Equal(t, "billing", typeDecl.Members[2].Spread.Ref.Name)
		require.Equal(t, "PaymentInfo", *typeDecl.Members[2].Spread.Ref.Member)

		// Regular field
		require.NotNil(t, typeDecl.Members[3].Field)
		require.Equal(t, "name", typeDecl.Members[3].Field.Name)
	})

	t.Run("Mixed local and namespaced spreads in enum body", func(t *testing.T) {
		input := `
			enum AllPermissions {
				...BasePermissions
				...auth.AdminPermissions
				Custom = "custom"
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		members := parsed.Declarations[0].Enum.Members
		require.Len(t, members, 3)

		require.NotNil(t, members[0].Spread)
		require.Equal(t, "BasePermissions", members[0].Spread.Ref.Name)
		require.Nil(t, members[0].Spread.Ref.Member)

		require.NotNil(t, members[1].Spread)
		require.Equal(t, "auth", members[1].Spread.Ref.Name)
		require.Equal(t, "AdminPermissions", *members[1].Spread.Ref.Member)

		require.Equal(t, "Custom", members[2].Name)
		require.Equal(t, "custom", members[2].Value.Str.String())
	})

	t.Run("Multiple namespaced spreads in const object", func(t *testing.T) {
		input := `
			const merged = {
				...defaults.base
				...overrides.prod
				name "final"
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		entries := parsed.Declarations[0].Const.Value.Object.Entries
		require.Len(t, entries, 3)

		require.NotNil(t, entries[0].Spread)
		require.Equal(t, "defaults", entries[0].Spread.Ref.Name)
		require.Equal(t, "base", *entries[0].Spread.Ref.Member)

		require.NotNil(t, entries[1].Spread)
		require.Equal(t, "overrides", entries[1].Spread.Ref.Name)
		require.Equal(t, "prod", *entries[1].Spread.Ref.Member)

		require.Equal(t, "name", entries[2].Key)
		require.Equal(t, "final", entries[2].Value.Scalar.Str.String())
	})

	t.Run("Namespaced spread in large schema with rpc-style type", func(t *testing.T) {
		input := `
			@rpc
			type UserService {
				...common.BaseService
				@proc
				GetUser {
					input {
						id string
					}
					output {
						...models.UserProfile
						active bool
					}
				}
			}
		`
		parsed, err := ParserInstance.ParseString("schema.vdl", input)
		require.NoError(t, err)

		typeDecl := parsed.Declarations[0].Type
		require.Equal(t, "UserService", typeDecl.Name)
		require.Len(t, typeDecl.Annotations, 1)
		require.Equal(t, "rpc", typeDecl.Annotations[0].Name)
		require.Len(t, typeDecl.Members, 2)

		// First child: namespaced spread
		require.NotNil(t, typeDecl.Members[0].Spread)
		require.Equal(t, "common", typeDecl.Members[0].Spread.Ref.Name)
		require.Equal(t, "BaseService", *typeDecl.Members[0].Spread.Ref.Member)

		// Second child: proc field with nested objects
		proc := typeDecl.Members[1].Field
		require.Equal(t, "GetUser", proc.Name)
		require.NotNil(t, proc.Type.Base.Object)
		require.Len(t, proc.Type.Base.Object.Members, 2)

		// output has a namespaced spread inside
		outputField := proc.Type.Base.Object.Members[1].Field
		require.Equal(t, "output", outputField.Name)
		require.NotNil(t, outputField.Type.Base.Object)
		outputChildren := outputField.Type.Base.Object.Members
		require.Len(t, outputChildren, 2)
		require.NotNil(t, outputChildren[0].Spread)
		require.Equal(t, "models", outputChildren[0].Spread.Ref.Name)
		require.Equal(t, "UserProfile", *outputChildren[0].Spread.Ref.Member)
		require.NotNil(t, outputChildren[1].Field)
		require.Equal(t, "active", outputChildren[1].Field.Name)
	})
}

func TestParserDataLiteralKeys(t *testing.T) {
	input := `
		const c = {
			type "a"
			enum "b"
			const "c"
			map "d"
			true "e"
			false "f"
		}
	`

	parsed, err := ParserInstance.ParseString("schema.vdl", input)
	require.NoError(t, err)

	entries := parsed.Declarations[0].Const.Value.Object.Entries
	require.Len(t, entries, 6)
	require.Equal(t, "type", entries[0].Key)
	require.Equal(t, "enum", entries[1].Key)
	require.Equal(t, "const", entries[2].Key)
	require.Equal(t, "map", entries[3].Key)
	require.Equal(t, "true", entries[4].Key)
	require.Equal(t, "false", entries[5].Key)
}
