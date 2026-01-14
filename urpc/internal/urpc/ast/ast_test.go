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
