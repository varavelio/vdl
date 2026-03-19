package testutil

import (
	"encoding/json"
	"fmt"
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

// IRJSONEqualNoPos compares two IR JSON payloads and ignores all nested
// `position` keys recursively.
func IRJSONEqualNoPos(t *testing.T, expectedJSON, actualJSON []byte, msgAndArgs ...any) {
	t.Helper()

	expectedClean, err := StripPositionsFromJSON(expectedJSON)
	require.NoError(t, err)

	actualClean, err := StripPositionsFromJSON(actualJSON)
	require.NoError(t, err)

	require.Equal(t, string(expectedClean), string(actualClean), msgAndArgs...)
}

// StripPositionsFromJSON removes all `position` keys recursively from a JSON payload.
func StripPositionsFromJSON(input []byte) ([]byte, error) {
	var data any
	if err := json.Unmarshal(input, &data); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}

	clean := stripPositionKeys(data)
	out, err := json.MarshalIndent(clean, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal cleaned json: %w", err)
	}

	return append(out, '\n'), nil
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

func stripPositionKeys(value any) any {
	switch v := value.(type) {
	case map[string]any:
		clean := make(map[string]any, len(v))
		for key, entry := range v {
			if key == "position" {
				continue
			}
			clean[key] = stripPositionKeys(entry)
		}
		return clean
	case []any:
		clean := make([]any, len(v))
		for i, entry := range v {
			clean[i] = stripPositionKeys(entry)
		}
		return clean
	default:
		return v
	}
}
