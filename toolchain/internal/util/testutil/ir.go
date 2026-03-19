package testutil

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/varavelio/vdl/toolchain/internal/core/ir/irtypes"
	"github.com/varavelio/vdl/toolchain/internal/util/debugutil"
)

// IRSchemaEqual compares two IR schemas and fails if they are not equal.
// The validation includes position fields.
func IRSchemaEqual(t *testing.T, expected, actual *irtypes.IrSchema, msgAndArgs ...any) {
	t.Helper()

	require.Equal(t, debugutil.ToBeautyJSON(expected), debugutil.ToBeautyJSON(actual), msgAndArgs...)
}

// IRSchemaEqualNoPos compares two IR schemas and fails if they are not equal.
// It ignores every nested Position field recursively.
func IRSchemaEqualNoPos(t *testing.T, expected, actual *irtypes.IrSchema, msgAndArgs ...any) {
	t.Helper()

	expectedCopy := cloneIRSchema(t, expected)
	actualCopy := cloneIRSchema(t, actual)

	irCleanPositionsRecursively(reflect.ValueOf(expectedCopy), reflect.ValueOf(irtypes.Position{}))
	irCleanPositionsRecursively(reflect.ValueOf(actualCopy), reflect.ValueOf(irtypes.Position{}))

	IRSchemaEqual(t, expectedCopy, actualCopy, msgAndArgs...)
}

func cloneIRSchema(t *testing.T, schema *irtypes.IrSchema) *irtypes.IrSchema {
	t.Helper()

	if schema == nil {
		return nil
	}

	data, err := json.Marshal(schema)
	require.NoError(t, err)

	var out irtypes.IrSchema
	require.NoError(t, json.Unmarshal(data, &out))

	return &out
}

func irCleanPositionsRecursively(val reflect.Value, emptyPos reflect.Value) {
	if !val.IsValid() {
		return
	}

	switch val.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !val.IsNil() {
			irCleanPositionsRecursively(val.Elem(), emptyPos)
		}
	case reflect.Struct:
		for _, field := range val.Fields() {
			if field.CanSet() && field.Type() == emptyPos.Type() {
				field.Set(emptyPos)
			}
			irCleanPositionsRecursively(field, emptyPos)
		}
	case reflect.Slice, reflect.Array:
		for i := range val.Len() {
			irCleanPositionsRecursively(val.Index(i), emptyPos)
		}
	}
}
