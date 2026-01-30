package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocstringGetExternal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		ok       bool
	}{
		{"valid unix path", "./docs/readme.md", "./docs/readme.md", true},
		{"valid absolute unix path", "/usr/local/readme.md", "/usr/local/readme.md", true},
		{"valid windows path", `C:\docs\readme.md`, `C:\docs\readme.md`, true},
		{"invalid extension", "./docs/readme.txt", "", false},
		{"empty string", "", "", false},
		{"only whitespace", "   ", "", false},
		{"newline at end", "./docs/readme.md\n", "./docs/readme.md", true},
		{"newline at start", "\n./docs/readme.md", "./docs/readme.md", true},
		{"newline in middle", "./docs/\nreadme.md", "", false},
		{"carriage return at end", "./docs/readme.md\r", "./docs/readme.md", true},
		{"carriage return in middle", "./docs/\rreadme.md", "", false},
		{"uppercase extension", "./docs/README.MD", "", false},
		{"leading and trailing whitespace", "  ./docs/readme.md  ", "./docs/readme.md", true},
		{"just .md", ".md", "", false},
		{"dotfile but not markdown", ".gitignore", "", false},
		{"directory with dot", "./.config/readme.md", "./.config/readme.md", true},
		{"tricky valid path with newline padding", "\n  ./dir/file.md  \r\n", "./dir/file.md", true},
		{"path with tab in middle", "./docs/\treadme.md", "./docs/\treadme.md", true},
	}

	for _, tt := range tests {
		d := Docstring{Value: DocstringValue(tt.input)}
		path, ok := d.GetExternal()
		require.Equal(t, tt.ok, ok, tt.name)
		require.Equal(t, tt.expected, path, tt.name)
	}
}

func TestDocstringIsExternal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		ok       bool
	}{
		{
			name:     "Valid file",
			input:    "external_doc.md",
			expected: "external_doc.md",
			ok:       true,
		},
		{
			name:     "With spaces",
			input:    "   doc.md   ",
			expected: "doc.md",
			ok:       true,
		},
		{
			name:     "With newline",
			input:    "some\ndoc.md",
			expected: "",
			ok:       false,
		},
		{
			name:     "Does not end with .md",
			input:    "doc.txt",
			expected: "",
			ok:       false,
		},
		{
			name:     "Incorrect suffix",
			input:    ".md",
			expected: "",
			ok:       false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := DocstringIsExternal(tt.input)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenedFields(t *testing.T) {
	primitiveString := "string"
	primitiveInt := "int"

	// Constructing a nested field structure for testing
	fURL := &Field{Name: "url", Type: FieldType{Base: &FieldTypeBase{Named: &primitiveString}}}
	fSize := &Field{Name: "size", Type: FieldType{Base: &FieldTypeBase{Named: &primitiveInt}}}
	fAvatar := &Field{
		Name: "avatar",
		Type: FieldType{
			Base: &FieldTypeBase{
				Object: &FieldTypeObject{
					Children: []*TypeDeclChild{
						{Field: fURL},
						{Field: fSize},
					},
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
					Children: []*TypeDeclChild{
						{Field: fName},
						{Field: fAvatar},
					},
				},
			},
		},
	}
	fId := &Field{Name: "id", Type: FieldType{Base: &FieldTypeBase{Named: &primitiveInt}}}
	fRole := &Field{Name: "role", Type: FieldType{Base: &FieldTypeBase{Named: &primitiveString}}}

	typeDeclChildren := []*TypeDeclChild{
		{Field: fId},
		{Comment: &Comment{Simple: &[]string{"// some comment"}[0]}},
		{Field: fProfile},
		{Field: fRole},
	}

	// Create InputOutputChild slice for Input/Output tests
	inputOutputChildren := []*InputOutputChild{
		{Field: fId},
		{Comment: &Comment{Simple: &[]string{"// some comment"}[0]}},
		{Field: fProfile},
		{Field: fRole},
	}

	expectedFieldNames := []string{"id", "profile", "name", "avatar", "url", "size", "role"}

	t.Run("TypeDecl", func(t *testing.T) {
		typeDecl := &TypeDecl{
			Name:     "User",
			Children: typeDeclChildren,
		}

		flattened := typeDecl.GetFlattenedFields()
		require.Len(t, flattened, len(expectedFieldNames), "should have correct number of fields")

		for i, field := range flattened {
			require.Equal(t, expectedFieldNames[i], field.Name, "field name should match")
		}

		// Test pointer modification
		require.False(t, flattened[0].Optional, "id should not be optional initially")
		flattened[0].Optional = true
		require.True(t, fId.Optional, "modifying flattened field should modify original field")
	})

	t.Run("ProcOrStreamDeclChildInput", func(t *testing.T) {
		inputDecl := &ProcOrStreamDeclChildInput{
			Children: inputOutputChildren,
		}

		flattened := inputDecl.GetFlattenedFields()
		require.Len(t, flattened, len(expectedFieldNames), "should have correct number of fields")

		for i, field := range flattened {
			require.Equal(t, expectedFieldNames[i], field.Name, "field name should match")
		}

		// Test pointer modification
		fId.Optional = false // reset from previous test
		require.False(t, flattened[0].Optional, "id should not be optional initially")
		flattened[0].Optional = true
		require.True(t, fId.Optional, "modifying flattened field should modify original field")
	})

	t.Run("ProcOrStreamDeclChildOutput", func(t *testing.T) {
		outputDecl := &ProcOrStreamDeclChildOutput{
			Children: inputOutputChildren,
		}

		flattened := outputDecl.GetFlattenedFields()
		require.Len(t, flattened, len(expectedFieldNames), "should have correct number of fields")

		for i, field := range flattened {
			require.Equal(t, expectedFieldNames[i], field.Name, "field name should match")
		}

		// Test pointer modification
		fId.Optional = false // reset from previous test
		require.False(t, flattened[0].Optional, "id should not be optional initially")
		flattened[0].Optional = true
		require.True(t, fId.Optional, "modifying flattened field should modify original field")
	})

	t.Run("Field.GetFlattenedField", func(t *testing.T) {
		flattened := fProfile.GetFlattenedField()
		expected := []string{"profile", "name", "avatar", "url", "size"}
		require.Len(t, flattened, len(expected), "should have correct number of fields")

		for i, field := range flattened {
			require.Equal(t, expected[i], field.Name, "field name should match")
		}

		// Test pointer modification
		require.False(t, flattened[1].Optional, "name should not be optional initially")
		flattened[1].Optional = true
		require.True(t, fName.Optional, "modifying flattened field should modify original field")
		fName.Optional = false // reset
	})
}

//////////////////////////////
// PRIMITIVE TYPE TESTS     //
//////////////////////////////

func TestIsPrimitiveType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"string is primitive", "string", true},
		{"int is primitive", "int", true},
		{"float is primitive", "float", true},
		{"bool is primitive", "bool", true},
		{"datetime is primitive", "datetime", true},
		{"custom type is not primitive", "User", false},
		{"empty string is not primitive", "", false},
		{"similar name is not primitive", "String", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPrimitiveType(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

//////////////////////////////
// CAPTURE TYPES TESTS      //
//////////////////////////////

func TestQuotedString(t *testing.T) {
	t.Run("Capture strips quotes", func(t *testing.T) {
		var qs QuotedString
		err := qs.Capture([]string{`"hello world"`})
		require.NoError(t, err)
		require.Equal(t, QuotedString("hello world"), qs)
	})

	t.Run("Capture handles empty string", func(t *testing.T) {
		var qs QuotedString
		err := qs.Capture([]string{`""`})
		require.NoError(t, err)
		require.Equal(t, QuotedString(""), qs)
	})

	t.Run("Capture without quotes", func(t *testing.T) {
		var qs QuotedString
		err := qs.Capture([]string{"no quotes"})
		require.NoError(t, err)
		require.Equal(t, QuotedString("no quotes"), qs)
	})

	t.Run("String method", func(t *testing.T) {
		qs := QuotedString("test value")
		require.Equal(t, "test value", qs.String())
	})
}

func TestDocstringValue(t *testing.T) {
	t.Run("Capture strips triple quotes", func(t *testing.T) {
		var dv DocstringValue
		err := dv.Capture([]string{`"""hello world"""`})
		require.NoError(t, err)
		require.Equal(t, DocstringValue("hello world"), dv)
	})

	t.Run("Capture handles empty docstring", func(t *testing.T) {
		var dv DocstringValue
		err := dv.Capture([]string{`""""""`})
		require.NoError(t, err)
		require.Equal(t, DocstringValue(""), dv)
	})

	t.Run("Capture handles multiline", func(t *testing.T) {
		var dv DocstringValue
		err := dv.Capture([]string{`"""
line 1
line 2
"""`})
		require.NoError(t, err)
		require.Equal(t, DocstringValue("\nline 1\nline 2\n"), dv)
	})

	t.Run("String method", func(t *testing.T) {
		dv := DocstringValue("test docstring")
		require.Equal(t, "test docstring", dv.String())
	})
}

//////////////////////////////
// SCHEMA GETTER TESTS      //
//////////////////////////////

// Helper function to create a comprehensive test schema
func createTestSchema() *Schema {
	ptr := func(s string) *string { return &s }

	return &Schema{
		Children: []*SchemaChild{
			// Includes
			{Include: &Include{Path: "./common.vdl"}},
			{Include: &Include{Path: "./auth.vdl"}},
			// Comments
			{Comment: &Comment{Simple: ptr("// A comment")}},
			{Comment: &Comment{Block: ptr("/* Block comment */")}},
			// Docstrings
			{Docstring: &Docstring{Value: " Standalone docstring "}},
			// Types
			{Type: &TypeDecl{Name: "User", Children: []*TypeDeclChild{}}},
			{Type: &TypeDecl{Name: "Product", Children: []*TypeDeclChild{}}},
			// Consts
			{Const: &ConstDecl{Name: "VERSION", Value: &ConstValue{Str: (*QuotedString)(ptr("1.0"))}}},
			{Const: &ConstDecl{Name: "MAX_SIZE", Value: &ConstValue{Int: ptr("100")}}},
			// Enums
			{Enum: &EnumDecl{Name: "Status", Members: []*EnumMember{{Name: "Active"}}}},
			{Enum: &EnumDecl{Name: "Priority", Members: []*EnumMember{{Name: "High"}}}},
			// Patterns
			{Pattern: &PatternDecl{Name: "CacheKey", Value: "cache:{id}"}},
			// RPCs
			{RPC: &RPCDecl{
				Name: "UserService",
				Children: []*RPCChild{
					{Proc: &ProcDecl{Name: "GetUser", Children: []*ProcOrStreamDeclChild{}}},
					{Proc: &ProcDecl{Name: "CreateUser", Children: []*ProcOrStreamDeclChild{}}},
					{Stream: &StreamDecl{Name: "UserUpdates", Children: []*ProcOrStreamDeclChild{}}},
				},
			}},
			{RPC: &RPCDecl{
				Name: "OrderService",
				Children: []*RPCChild{
					{Proc: &ProcDecl{Name: "GetOrder", Children: []*ProcOrStreamDeclChild{}}},
					{Stream: &StreamDecl{Name: "OrderUpdates", Children: []*ProcOrStreamDeclChild{}}},
					{Stream: &StreamDecl{Name: "OrderNotifications", Children: []*ProcOrStreamDeclChild{}}},
				},
			}},
		},
	}
}

func TestSchemaGetIncludes(t *testing.T) {
	schema := createTestSchema()
	includes := schema.GetIncludes()

	require.Len(t, includes, 2)
	require.Equal(t, QuotedString("./common.vdl"), includes[0].Path)
	require.Equal(t, QuotedString("./auth.vdl"), includes[1].Path)
}

func TestSchemaGetComments(t *testing.T) {
	schema := createTestSchema()
	comments := schema.GetComments()

	require.Len(t, comments, 2)
	require.NotNil(t, comments[0].Simple)
	require.Equal(t, "// A comment", *comments[0].Simple)
	require.NotNil(t, comments[1].Block)
	require.Equal(t, "/* Block comment */", *comments[1].Block)
}

func TestSchemaGetDocstrings(t *testing.T) {
	schema := createTestSchema()
	docstrings := schema.GetDocstrings()

	require.Len(t, docstrings, 1)
	require.Equal(t, DocstringValue(" Standalone docstring "), docstrings[0].Value)
}

func TestSchemaGetTypes(t *testing.T) {
	schema := createTestSchema()
	types := schema.GetTypes()

	require.Len(t, types, 2)
	require.Equal(t, "User", types[0].Name)
	require.Equal(t, "Product", types[1].Name)
}

func TestSchemaGetTypesMap(t *testing.T) {
	schema := createTestSchema()
	typesMap := schema.GetTypesMap()

	require.Len(t, typesMap, 2)
	require.NotNil(t, typesMap["User"])
	require.NotNil(t, typesMap["Product"])
	require.Nil(t, typesMap["NonExistent"])
	require.Equal(t, "User", typesMap["User"].Name)
}

func TestSchemaGetConsts(t *testing.T) {
	schema := createTestSchema()
	consts := schema.GetConsts()

	require.Len(t, consts, 2)
	require.Equal(t, "VERSION", consts[0].Name)
	require.Equal(t, "MAX_SIZE", consts[1].Name)
}

func TestSchemaGetEnums(t *testing.T) {
	schema := createTestSchema()
	enums := schema.GetEnums()

	require.Len(t, enums, 2)
	require.Equal(t, "Status", enums[0].Name)
	require.Equal(t, "Priority", enums[1].Name)
}

func TestSchemaGetPatterns(t *testing.T) {
	schema := createTestSchema()
	patterns := schema.GetPatterns()

	require.Len(t, patterns, 1)
	require.Equal(t, "CacheKey", patterns[0].Name)
	require.Equal(t, QuotedString("cache:{id}"), patterns[0].Value)
}

func TestSchemaGetRPCs(t *testing.T) {
	schema := createTestSchema()
	rpcs := schema.GetRPCs()

	require.Len(t, rpcs, 2)
	require.Equal(t, "UserService", rpcs[0].Name)
	require.Equal(t, "OrderService", rpcs[1].Name)
}

func TestSchemaGetRPCsMap(t *testing.T) {
	schema := createTestSchema()
	rpcsMap := schema.GetRPCsMap()

	require.Len(t, rpcsMap, 2)
	require.NotNil(t, rpcsMap["UserService"])
	require.NotNil(t, rpcsMap["OrderService"])
	require.Nil(t, rpcsMap["NonExistent"])
	require.Equal(t, "UserService", rpcsMap["UserService"].Name)
}

func TestSchemaEmptySchema(t *testing.T) {
	schema := &Schema{Children: []*SchemaChild{}}

	require.Empty(t, schema.GetIncludes())
	require.Empty(t, schema.GetComments())
	require.Empty(t, schema.GetDocstrings())
	require.Empty(t, schema.GetTypes())
	require.Empty(t, schema.GetTypesMap())
	require.Empty(t, schema.GetConsts())
	require.Empty(t, schema.GetEnums())
	require.Empty(t, schema.GetPatterns())
	require.Empty(t, schema.GetRPCs())
	require.Empty(t, schema.GetRPCsMap())
}

//////////////////////////////
// SCHEMA CHILD KIND TESTS  //
//////////////////////////////

func TestSchemaChildKind(t *testing.T) {
	ptr := func(s string) *string { return &s }

	tests := []struct {
		name     string
		child    *SchemaChild
		expected SchemaChildKind
	}{
		{"Include", &SchemaChild{Include: &Include{Path: "./test.vdl"}}, SchemaChildKindInclude},
		{"Comment", &SchemaChild{Comment: &Comment{Simple: ptr("// comment")}}, SchemaChildKindComment},
		{"Docstring", &SchemaChild{Docstring: &Docstring{Value: "doc"}}, SchemaChildKindDocstring},
		{"Type", &SchemaChild{Type: &TypeDecl{Name: "User"}}, SchemaChildKindType},
		{"Const", &SchemaChild{Const: &ConstDecl{Name: "MAX"}}, SchemaChildKindConst},
		{"Enum", &SchemaChild{Enum: &EnumDecl{Name: "Status"}}, SchemaChildKindEnum},
		{"Pattern", &SchemaChild{Pattern: &PatternDecl{Name: "Key"}}, SchemaChildKindPattern},
		{"RPC", &SchemaChild{RPC: &RPCDecl{Name: "Service"}}, SchemaChildKindRPC},
		{"Empty", &SchemaChild{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.child.Kind())
		})
	}
}

//////////////////////////////
// RPC GETTER TESTS         //
//////////////////////////////

func TestRPCDeclGetProcs(t *testing.T) {
	rpc := &RPCDecl{
		Name: "TestService",
		Children: []*RPCChild{
			{Proc: &ProcDecl{Name: "Proc1"}},
			{Comment: &Comment{Simple: func() *string { s := "// comment"; return &s }()}},
			{Proc: &ProcDecl{Name: "Proc2"}},
			{Stream: &StreamDecl{Name: "Stream1"}},
			{Proc: &ProcDecl{Name: "Proc3"}},
		},
	}

	procs := rpc.GetProcs()

	require.Len(t, procs, 3)
	require.Equal(t, "Proc1", procs[0].Name)
	require.Equal(t, "Proc2", procs[1].Name)
	require.Equal(t, "Proc3", procs[2].Name)
}

func TestRPCDeclGetStreams(t *testing.T) {
	rpc := &RPCDecl{
		Name: "TestService",
		Children: []*RPCChild{
			{Stream: &StreamDecl{Name: "Stream1"}},
			{Proc: &ProcDecl{Name: "Proc1"}},
			{Stream: &StreamDecl{Name: "Stream2"}},
			{Comment: &Comment{Simple: func() *string { s := "// comment"; return &s }()}},
			{Stream: &StreamDecl{Name: "Stream3"}},
		},
	}

	streams := rpc.GetStreams()

	require.Len(t, streams, 3)
	require.Equal(t, "Stream1", streams[0].Name)
	require.Equal(t, "Stream2", streams[1].Name)
	require.Equal(t, "Stream3", streams[2].Name)
}

func TestRPCDeclEmptyRPC(t *testing.T) {
	rpc := &RPCDecl{
		Name:     "EmptyService",
		Children: []*RPCChild{},
	}

	require.Empty(t, rpc.GetProcs())
	require.Empty(t, rpc.GetStreams())
}

//////////////////////////////
// CONST VALUE TESTS        //
//////////////////////////////

func TestConstValueString(t *testing.T) {
	ptr := func(s string) *string { return &s }
	qptr := func(s string) *QuotedString { q := QuotedString(s); return &q }

	tests := []struct {
		name     string
		value    ConstValue
		expected string
	}{
		{"String value", ConstValue{Str: qptr("hello")}, `"hello"`},
		{"String with quotes", ConstValue{Str: qptr(`say "hi"`)}, `"say \"hi\""`},
		{"Integer value", ConstValue{Int: ptr("42")}, "42"},
		{"Float value", ConstValue{Float: ptr("3.14")}, "3.14"},
		{"Boolean true", ConstValue{True: true}, "true"},
		{"Boolean false", ConstValue{False: true}, "false"},
		{"Empty value", ConstValue{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.value.String())
		})
	}
}

//////////////////////////////
// ANY LITERAL TESTS        //
//////////////////////////////

func TestAnyLiteralString(t *testing.T) {
	ptr := func(s string) *string { return &s }
	qptr := func(s string) *QuotedString { q := QuotedString(s); return &q }

	tests := []struct {
		name     string
		value    AnyLiteral
		expected string
	}{
		{"String value", AnyLiteral{Str: qptr("hello")}, `"hello"`},
		{"String with quotes", AnyLiteral{Str: qptr(`say "hi"`)}, `"say \"hi\""`},
		{"Integer value", AnyLiteral{Int: ptr("42")}, "42"},
		{"Float value", AnyLiteral{Float: ptr("3.14")}, "3.14"},
		{"Boolean true", AnyLiteral{True: true}, "true"},
		{"Boolean false", AnyLiteral{False: true}, "false"},
		{"Empty value", AnyLiteral{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.value.String())
		})
	}
}

//////////////////////////////
// FIELD TYPE ARRAY TESTS   //
//////////////////////////////

func TestFieldTypeIsArray(t *testing.T) {
	ptr := func(s string) *string { return &s }

	tests := []struct {
		name     string
		ft       FieldType
		expected bool
	}{
		{
			name:     "Non-array type",
			ft:       FieldType{Base: &FieldTypeBase{Named: ptr("string")}},
			expected: false,
		},
		{
			name:     "1D array",
			ft:       FieldType{Base: &FieldTypeBase{Named: ptr("string")}, Dimensions: 1},
			expected: true,
		},
		{
			name:     "2D array",
			ft:       FieldType{Base: &FieldTypeBase{Named: ptr("int")}, Dimensions: 2},
			expected: true,
		},
		{
			name:     "3D array",
			ft:       FieldType{Base: &FieldTypeBase{Named: ptr("float")}, Dimensions: 3},
			expected: true,
		},
		{
			name:     "Zero dimensions",
			ft:       FieldType{Base: &FieldTypeBase{Named: ptr("bool")}, Dimensions: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.ft.IsArray())
		})
	}
}

func TestFieldTypeDimensions(t *testing.T) {
	ptr := func(s string) *string { return &s }

	tests := []struct {
		name     string
		ft       FieldType
		expected ArrayDimensions
	}{
		{
			name:     "Non-array type",
			ft:       FieldType{Base: &FieldTypeBase{Named: ptr("string")}},
			expected: 0,
		},
		{
			name:     "1D array",
			ft:       FieldType{Base: &FieldTypeBase{Named: ptr("string")}, Dimensions: 1},
			expected: 1,
		},
		{
			name:     "2D array",
			ft:       FieldType{Base: &FieldTypeBase{Named: ptr("int")}, Dimensions: 2},
			expected: 2,
		},
		{
			name:     "3D array",
			ft:       FieldType{Base: &FieldTypeBase{Named: ptr("float")}, Dimensions: 3},
			expected: 3,
		},
		{
			name:     "5D array",
			ft:       FieldType{Base: &FieldTypeBase{Named: ptr("datetime")}, Dimensions: 5},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.ft.Dimensions)
		})
	}
}

func TestArrayDimensionsCapture(t *testing.T) {
	t.Run("No calls means zero", func(t *testing.T) {
		var ad ArrayDimensions
		require.Equal(t, ArrayDimensions(0), ad)
	})

	t.Run("Single capture call", func(t *testing.T) {
		var ad ArrayDimensions
		err := ad.Capture([]string{"]"})
		require.NoError(t, err)
		require.Equal(t, ArrayDimensions(1), ad)
	})

	t.Run("Multiple capture calls accumulate", func(t *testing.T) {
		var ad ArrayDimensions
		// participle calls Capture once per match in a repetition
		_ = ad.Capture([]string{"]"}) // first []
		_ = ad.Capture([]string{"]"}) // second []
		require.Equal(t, ArrayDimensions(2), ad)
	})

	t.Run("Three dimensions", func(t *testing.T) {
		var ad ArrayDimensions
		_ = ad.Capture([]string{"]"})
		_ = ad.Capture([]string{"]"})
		_ = ad.Capture([]string{"]"})
		require.Equal(t, ArrayDimensions(3), ad)
	})
}
