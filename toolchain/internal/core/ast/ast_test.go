package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocstringIsExternal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		ok       bool
	}{
		{name: "valid markdown path", input: " ./docs/readme.md ", expected: "./docs/readme.md", ok: true},
		{name: "valid windows path", input: `C:\docs\readme.md`, expected: `C:\docs\readme.md`, ok: true},
		{name: "invalid when multiline", input: "./docs\nreadme.md", expected: "", ok: false},
		{name: "invalid extension", input: "./docs/readme.txt", expected: "", ok: false},
		{name: "uppercase extension invalid", input: "./docs/README.MD", expected: "", ok: false},
		{name: "empty invalid", input: "", expected: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := DocstringIsExternal(tt.input)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPrimitiveType(t *testing.T) {
	require.True(t, IsPrimitiveType("string"))
	require.True(t, IsPrimitiveType("datetime"))
	require.False(t, IsPrimitiveType("User"))
}

func TestSchemaGetters(t *testing.T) {
	schema := &Schema{
		Declarations: []*TopLevelDecl{
			{Include: &Include{Path: "./common.vdl"}},
			{Docstring: &Docstring{Value: " standalone "}},
			{Type: &TypeDecl{Name: "User"}},
			{Const: &ConstDecl{Name: "AppConfig"}},
			{Enum: &EnumDecl{Name: "Role"}},
		},
	}

	require.Len(t, schema.GetIncludes(), 1)
	require.Len(t, schema.GetDocstrings(), 1)
	require.Len(t, schema.GetTypes(), 1)
	require.Len(t, schema.GetConsts(), 1)
	require.Len(t, schema.GetEnums(), 1)
	require.Equal(t, "User", schema.GetTypesMap()["User"].Name)
	require.Nil(t, schema.GetTypesMap()["Unknown"])
}

func TestSchemaChildKind(t *testing.T) {
	tests := []struct {
		name     string
		child    *TopLevelDecl
		expected DeclKind
	}{
		{name: "include", child: &TopLevelDecl{Include: &Include{Path: "./foo.vdl"}}, expected: DeclKindInclude},
		{name: "docstring", child: &TopLevelDecl{Docstring: &Docstring{Value: " doc "}}, expected: DeclKindDocstring},
		{name: "type", child: &TopLevelDecl{Type: &TypeDecl{Name: "User"}}, expected: DeclKindType},
		{name: "const", child: &TopLevelDecl{Const: &ConstDecl{Name: "Config"}}, expected: DeclKindConst},
		{name: "enum", child: &TopLevelDecl{Enum: &EnumDecl{Name: "Role"}}, expected: DeclKindEnum},
		{name: "empty", child: &TopLevelDecl{}, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.child.Kind())
		})
	}
}

func TestFlattenedFields(t *testing.T) {
	primitiveString := "string"
	primitiveInt := "int"

	fURL := &Field{Name: "url", Type: FieldType{Base: &FieldTypeBase{Named: &primitiveString}}}
	fSize := &Field{Name: "size", Type: FieldType{Base: &FieldTypeBase{Named: &primitiveInt}}}
	fAvatar := &Field{
		Name: "avatar",
		Type: FieldType{
			Base: &FieldTypeBase{
				Object: &FieldTypeObject{
					Members: []*TypeMember{{Field: fURL}, {Field: fSize}},
				},
			},
		},
	}
	fName := &Field{Name: "name", Type: FieldType{Base: &FieldTypeBase{Named: &primitiveString}}}
	fProfile := &Field{
		Name: "profile",
		Type: FieldType{
			Base: &FieldTypeBase{
				Object: &FieldTypeObject{
					Members: []*TypeMember{{Field: fName}, {Field: fAvatar}},
				},
			},
		},
	}

	typeDecl := &TypeDecl{
		Name: "User",
		Members: []*TypeMember{
			{Field: &Field{Name: "id", Type: FieldType{Base: &FieldTypeBase{Named: &primitiveInt}}}},
			{Field: fProfile},
		},
	}

	flattened := typeDecl.GetFlattenedFields()
	require.Len(t, flattened, 6)
	require.Equal(t, "id", flattened[0].Name)
	require.Equal(t, "profile", flattened[1].Name)
	require.Equal(t, "name", flattened[2].Name)
	require.Equal(t, "avatar", flattened[3].Name)
	require.Equal(t, "url", flattened[4].Name)
	require.Equal(t, "size", flattened[5].Name)
}

func TestAnyLiteralString(t *testing.T) {
	ptr := func(s string) *string { return &s }
	qptr := func(s string) *QuotedString { q := QuotedString(s); return &q }

	require.Equal(t, `"hello"`, ScalarLiteral{Str: qptr("hello")}.String())
	require.Equal(t, "42", ScalarLiteral{Int: ptr("42")}.String())
	require.Equal(t, "3.14", ScalarLiteral{Float: ptr("3.14")}.String())
	require.Equal(t, "true", ScalarLiteral{True: true}.String())
	require.Equal(t, "false", ScalarLiteral{False: true}.String())
	require.Equal(t, "", ScalarLiteral{}.String())
}

func TestCaptureTypes(t *testing.T) {
	t.Run("quoted string capture", func(t *testing.T) {
		var qs QuotedString
		require.NoError(t, qs.Capture([]string{`"hello"`}))
		require.Equal(t, QuotedString("hello"), qs)
		require.Equal(t, "hello", qs.String())
	})

	t.Run("docstring capture", func(t *testing.T) {
		var dv DocstringValue
		require.NoError(t, dv.Capture([]string{`"""hello"""`}))
		require.Equal(t, DocstringValue("hello"), dv)
		require.Equal(t, "hello", dv.String())
	})
}

func TestFieldTypeIsArray(t *testing.T) {
	ptr := func(s string) *string { return &s }

	require.False(t, (&FieldType{Base: &FieldTypeBase{Named: ptr("string")}}).IsArray())
	require.True(t, (&FieldType{Base: &FieldTypeBase{Named: ptr("string")}, Dimensions: 2}).IsArray())
}

func TestArrayDimensionsCapture(t *testing.T) {
	var ad ArrayDimensions
	require.NoError(t, ad.Capture([]string{"]"}))
	require.NoError(t, ad.Capture([]string{"]"}))
	require.Equal(t, ArrayDimensions(2), ad)
	require.NoError(t, ad.Capture([]string{"]"}))
	require.Equal(t, ArrayDimensions(3), ad)
}

func TestReferenceString(t *testing.T) {
	t.Run("simple const reference", func(t *testing.T) {
		ref := Reference{Name: "FOO"}
		require.Equal(t, "FOO", ref.String())
	})

	t.Run("enum member reference", func(t *testing.T) {
		member := "Red"
		ref := Reference{Name: "Color", Member: &member}
		require.Equal(t, "Color.Red", ref.String())
	})

	t.Run("nil member is simple reference", func(t *testing.T) {
		ref := Reference{Name: "BASE_CONFIG", Member: nil}
		require.Equal(t, "BASE_CONFIG", ref.String())
	})
}

func TestAnyLiteralStringWithRef(t *testing.T) {
	t.Run("ref without member", func(t *testing.T) {
		ref := &Reference{Name: "MY_CONST"}
		lit := ScalarLiteral{Ref: ref}
		require.Equal(t, "MY_CONST", lit.String())
	})

	t.Run("ref with member", func(t *testing.T) {
		member := "Active"
		ref := &Reference{Name: "Status", Member: &member}
		lit := ScalarLiteral{Ref: ref}
		require.Equal(t, "Status.Active", lit.String())
	})

	t.Run("ref takes lowest precedence", func(t *testing.T) {
		// When Str is set, Ref should be ignored (Str is checked first)
		q := QuotedString("hello")
		member := "X"
		lit := ScalarLiteral{
			Str: &q,
			Ref: &Reference{Name: "A", Member: &member},
		}
		require.Equal(t, `"hello"`, lit.String())
	})
}

func TestScalarLiteralStringAllTypes(t *testing.T) {
	ptr := func(s string) *string { return &s }
	qptr := func(s string) *QuotedString { q := QuotedString(s); return &q }

	tests := []struct {
		name     string
		literal  ScalarLiteral
		expected string
	}{
		{name: "string literal", literal: ScalarLiteral{Str: qptr("hello")}, expected: `"hello"`},
		{name: "string with quotes", literal: ScalarLiteral{Str: qptr(`say "hi"`)}, expected: `"say \"hi\""`},
		{name: "int literal", literal: ScalarLiteral{Int: ptr("42")}, expected: "42"},
		{name: "float literal", literal: ScalarLiteral{Float: ptr("3.14")}, expected: "3.14"},
		{name: "true literal", literal: ScalarLiteral{True: true}, expected: "true"},
		{name: "false literal", literal: ScalarLiteral{False: true}, expected: "false"},
		{name: "simple ref", literal: ScalarLiteral{Ref: &Reference{Name: "FOO"}}, expected: "FOO"},
		{name: "enum ref", literal: ScalarLiteral{Ref: &Reference{Name: "Color", Member: ptr("Red")}}, expected: "Color.Red"},
		{name: "empty literal", literal: ScalarLiteral{}, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.literal.String())
		})
	}
}

func TestTypeDeclChildWithDocstring(t *testing.T) {
	t.Run("docstring child", func(t *testing.T) {
		child := &TypeMember{Docstring: &Docstring{Value: " field docs "}}
		require.NotNil(t, child.Docstring)
		require.Nil(t, child.Field)
		require.Nil(t, child.Spread)
	})

	t.Run("field child", func(t *testing.T) {
		child := &TypeMember{Field: &Field{Name: "id"}}
		require.Nil(t, child.Docstring)
		require.NotNil(t, child.Field)
		require.Nil(t, child.Spread)
	})

	t.Run("spread child", func(t *testing.T) {
		child := &TypeMember{Spread: &Spread{Ref: &Reference{Name: "Base"}}}
		require.Nil(t, child.Docstring)
		require.Nil(t, child.Field)
		require.NotNil(t, child.Spread)
	})
}

func TestSpreadWithReference(t *testing.T) {
	t.Run("spread with simple reference", func(t *testing.T) {
		s := &Spread{Ref: &Reference{Name: "BaseUser"}}
		require.NotNil(t, s.Ref)
		require.Equal(t, "BaseUser", s.Ref.Name)
		require.Nil(t, s.Ref.Member)
		require.Equal(t, "BaseUser", s.Ref.String())
	})

	t.Run("spread with namespaced reference", func(t *testing.T) {
		member := "BaseUser"
		s := &Spread{Ref: &Reference{Name: "auth", Member: &member}}
		require.NotNil(t, s.Ref)
		require.Equal(t, "auth", s.Ref.Name)
		require.NotNil(t, s.Ref.Member)
		require.Equal(t, "BaseUser", *s.Ref.Member)
		require.Equal(t, "auth.BaseUser", s.Ref.String())
	})

	t.Run("spread in type decl child with simple ref", func(t *testing.T) {
		child := &TypeMember{Spread: &Spread{Ref: &Reference{Name: "Metadata"}}}
		require.NotNil(t, child.Spread)
		require.Equal(t, "Metadata", child.Spread.Ref.Name)
		require.Nil(t, child.Spread.Ref.Member)
	})

	t.Run("spread in type decl child with namespaced ref", func(t *testing.T) {
		member := "Credentials"
		child := &TypeMember{Spread: &Spread{Ref: &Reference{Name: "auth", Member: &member}}}
		require.NotNil(t, child.Spread)
		require.Equal(t, "auth", child.Spread.Ref.Name)
		require.Equal(t, "Credentials", *child.Spread.Ref.Member)
		require.Equal(t, "auth.Credentials", child.Spread.Ref.String())
	})

	t.Run("spread in enum member with simple ref", func(t *testing.T) {
		m := &EnumMember{Spread: &Spread{Ref: &Reference{Name: "BaseRoles"}}}
		require.NotNil(t, m.Spread)
		require.Equal(t, "BaseRoles", m.Spread.Ref.Name)
		require.Nil(t, m.Spread.Ref.Member)
	})

	t.Run("spread in enum member with namespaced ref", func(t *testing.T) {
		member := "AdminRoles"
		m := &EnumMember{Spread: &Spread{Ref: &Reference{Name: "auth", Member: &member}}}
		require.NotNil(t, m.Spread)
		require.Equal(t, "auth", m.Spread.Ref.Name)
		require.Equal(t, "AdminRoles", *m.Spread.Ref.Member)
		require.Equal(t, "auth.AdminRoles", m.Spread.Ref.String())
	})

	t.Run("spread in data literal object entry with simple ref", func(t *testing.T) {
		entry := &DataLiteralObjectEntry{
			Spread: &Spread{Ref: &Reference{Name: "baseConfig"}},
		}
		require.NotNil(t, entry.Spread)
		require.Equal(t, "baseConfig", entry.Spread.Ref.Name)
		require.Nil(t, entry.Spread.Ref.Member)
	})

	t.Run("spread in data literal object entry with namespaced ref", func(t *testing.T) {
		member := "prodDefaults"
		entry := &DataLiteralObjectEntry{
			Spread: &Spread{Ref: &Reference{Name: "config", Member: &member}},
		}
		require.NotNil(t, entry.Spread)
		require.Equal(t, "config", entry.Spread.Ref.Name)
		require.Equal(t, "prodDefaults", *entry.Spread.Ref.Member)
		require.Equal(t, "config.prodDefaults", entry.Spread.Ref.String())
	})
}

func TestDocstringGetExternalMethod(t *testing.T) {
	d := Docstring{Value: DocstringValue(" ./guide.md ")}
	path, ok := d.GetExternal()
	require.True(t, ok)
	require.Equal(t, "./guide.md", path)
}

func TestEnumMemberStructure(t *testing.T) {
	qptr := func(s string) *QuotedString { q := QuotedString(s); return &q }
	ptr := func(s string) *string { return &s }

	t.Run("plain member", func(t *testing.T) {
		m := &EnumMember{Name: "Active"}
		require.Equal(t, "Active", m.Name)
		require.Nil(t, m.Spread)
		require.Nil(t, m.Docstring)
		require.Len(t, m.Annotations, 0)
		require.Nil(t, m.Value)
	})

	t.Run("member with string value", func(t *testing.T) {
		m := &EnumMember{
			Name:  "Admin",
			Value: &EnumValue{Str: qptr("admin")},
		}
		require.Equal(t, "Admin", m.Name)
		require.NotNil(t, m.Value)
		require.Equal(t, "admin", m.Value.Str.String())
	})

	t.Run("member with int value", func(t *testing.T) {
		m := &EnumMember{
			Name:  "High",
			Value: &EnumValue{Int: ptr("10")},
		}
		require.Equal(t, "High", m.Name)
		require.NotNil(t, m.Value)
		require.Equal(t, "10", *m.Value.Int)
	})

	t.Run("spread member", func(t *testing.T) {
		m := &EnumMember{Spread: &Spread{Ref: &Reference{Name: "BaseRoles"}}}
		require.NotNil(t, m.Spread)
		require.Equal(t, "BaseRoles", m.Spread.Ref.Name)
		require.Equal(t, "", m.Name)
	})

	t.Run("member with flag annotation", func(t *testing.T) {
		m := &EnumMember{
			Annotations: []*Annotation{{Name: "deprecated"}},
			Name:        "OldValue",
		}
		require.Len(t, m.Annotations, 1)
		require.Equal(t, "deprecated", m.Annotations[0].Name)
		require.Nil(t, m.Annotations[0].Argument)
	})

	t.Run("member with annotation payload", func(t *testing.T) {
		m := &EnumMember{
			Annotations: []*Annotation{{
				Name:     "deprecated",
				Argument: &DataLiteral{Scalar: &ScalarLiteral{Str: qptr("Use NewValue")}},
			}},
			Name:  "OldValue",
			Value: &EnumValue{Str: qptr("old")},
		}
		require.Len(t, m.Annotations, 1)
		require.Equal(t, "deprecated", m.Annotations[0].Name)
		require.Equal(t, "Use NewValue", m.Annotations[0].Argument.Scalar.Str.String())
		require.Equal(t, "OldValue", m.Name)
		require.Equal(t, "old", m.Value.Str.String())
	})

	t.Run("member with multiple annotations", func(t *testing.T) {
		m := &EnumMember{
			Annotations: []*Annotation{
				{Name: "deprecated", Argument: &DataLiteral{Scalar: &ScalarLiteral{Str: qptr("Use X")}}},
				{Name: "alias", Argument: &DataLiteral{Scalar: &ScalarLiteral{Str: qptr("old_val")}}},
			},
			Name: "Legacy",
		}
		require.Len(t, m.Annotations, 2)
		require.Equal(t, "deprecated", m.Annotations[0].Name)
		require.Equal(t, "alias", m.Annotations[1].Name)
	})

	t.Run("member with docstring", func(t *testing.T) {
		m := &EnumMember{
			Docstring: &Docstring{Value: " The active state "},
			Name:      "Active",
		}
		require.NotNil(t, m.Docstring)
		require.Equal(t, " The active state ", m.Docstring.Value.String())
		require.Equal(t, "Active", m.Name)
	})

	t.Run("member with docstring and annotations", func(t *testing.T) {
		m := &EnumMember{
			Docstring:   &Docstring{Value: " Full access role "},
			Annotations: []*Annotation{{Name: "rbac", Argument: &DataLiteral{Scalar: &ScalarLiteral{Int: ptr("100")}}}},
			Name:        "Admin",
			Value:       &EnumValue{Str: qptr("admin")},
		}
		require.NotNil(t, m.Docstring)
		require.Equal(t, " Full access role ", m.Docstring.Value.String())
		require.Len(t, m.Annotations, 1)
		require.Equal(t, "rbac", m.Annotations[0].Name)
		require.Equal(t, "100", *m.Annotations[0].Argument.Scalar.Int)
		require.Equal(t, "Admin", m.Name)
		require.Equal(t, "admin", m.Value.Str.String())
	})

	t.Run("member with annotation with object payload", func(t *testing.T) {
		m := &EnumMember{
			Annotations: []*Annotation{{
				Name: "meta",
				Argument: &DataLiteral{Object: &DataLiteralObject{Entries: []*DataLiteralObjectEntry{
					{Key: "level", Value: &DataLiteral{Scalar: &ScalarLiteral{Int: ptr("100")}}},
					{Key: "label", Value: &DataLiteral{Scalar: &ScalarLiteral{Str: qptr("Super Admin")}}},
				}}},
			}},
			Name: "SuperAdmin",
		}
		require.Len(t, m.Annotations, 1)
		ann := m.Annotations[0]
		require.Equal(t, "meta", ann.Name)
		require.NotNil(t, ann.Argument.Object)
		require.Len(t, ann.Argument.Object.Entries, 2)
		require.Equal(t, "level", ann.Argument.Object.Entries[0].Key)
		require.Equal(t, "label", ann.Argument.Object.Entries[1].Key)
	})
}
