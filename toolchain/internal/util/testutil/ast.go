package testutil

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uforg/uforpc/urpc/internal/core/ast"
	"github.com/uforg/uforpc/urpc/internal/util/debugutil"
)

///////////////////////////
// AST Testing utilities //
///////////////////////////

// astCleanPositionsRecursively cleans all position fields recursively in any struct or array of structs.
// If includeRoot is true, it will also clean the position fields of the root object.
func astCleanPositionsRecursively(val reflect.Value, emptyPos reflect.Value, includeRoot bool) {
	if !val.IsValid() {
		return
	}

	switch val.Kind() {
	case reflect.Ptr:
		if !val.IsNil() {
			// Skip cleaning root if includeRoot is false
			astCleanPositionsRecursively(val.Elem(), emptyPos, includeRoot)
		}
	case reflect.Struct:
		// Set Pos and EndPos fields to empty value if they exist and we should process this level
		if includeRoot {
			if f := val.FieldByName("Pos"); f.IsValid() && f.CanSet() && f.Type() == emptyPos.Type() {
				f.Set(emptyPos)
			}
			if f := val.FieldByName("EndPos"); f.IsValid() && f.CanSet() && f.Type() == emptyPos.Type() {
				f.Set(emptyPos)
			}
		}

		// Always process fields recursively - after processing the current level
		for i := range val.NumField() {
			field := val.Field(i)
			if field.CanInterface() {
				// Always clean position fields in children
				astCleanPositionsRecursively(field, emptyPos, true)
			}
		}
	case reflect.Slice:
		// Handle arrays/slices recursively
		for i := range val.Len() {
			astCleanPositionsRecursively(val.Index(i), emptyPos, true)
		}
	}
}

// ASTEqual compares two URPC structs and fails if they are not ASTEqual.
// The validation includes the positions of the AST nodes.
func ASTEqual(t *testing.T, expected, actual *ast.Schema, msgAndArgs ...any) {
	t.Helper()

	// Prettify the JSON for better test error messages and diffs
	expectedJSON := debugutil.ToBeautyJSON(expected)
	actualJSON := debugutil.ToBeautyJSON(actual)

	require.Equal(t, expectedJSON, actualJSON, msgAndArgs...)
}

// ASTEqualNoPos compares two URPC structs and fails if they are not equal.
// It ignores the positions of any nested AST nodes.
func ASTEqualNoPos(t *testing.T, expected, actual *ast.Schema, msgAndArgs ...any) {
	t.Helper()

	cleanPositions := func(schema *ast.Schema) *ast.Schema {
		// Make a deep copy to avoid modifying the original
		schemaCopy := &ast.Schema{}
		*schemaCopy = *schema

		// Recursively clean all positions in the copy
		schemaVal := reflect.ValueOf(schemaCopy)
		astCleanPositionsRecursively(schemaVal, reflect.ValueOf(ast.Position{}), true)

		return schemaCopy
	}

	expected = cleanPositions(expected)
	actual = cleanPositions(actual)
	ASTEqual(t, expected, actual, msgAndArgs...)
}
