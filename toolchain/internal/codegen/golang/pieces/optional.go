package pieces

import (
	"encoding/json"
)

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Optional utility type
// -----------------------------------------------------------------------------

// Optional represents a general purpose value that may or may not be present.
//
// Can be used for handling nullable JSON fields.
//
// It is generic and works with any type T.
type Optional[T any] struct {
	Present bool // True if the value is present; false otherwise.
	Value   T    // The actual value when present; otherwise, the zero value of T.
}

// Some creates an Optional[T] with a present value set to v.
func Some[T any](v T) Optional[T] {
	return Optional[T]{Value: v, Present: true}
}

// None creates an Optional[T] with no present value, using the zero value of T.
func None[T any]() Optional[T] {
	return Optional[T]{Present: false}
}

// UnmarshalJSON implements json.Unmarshaler.
//
// It sets the Optional to absent if the JSON is "null"; otherwise, it unmarshals the value and marks it as present.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	// same as string(data) == "null" but without allocation
	if len(data) == 4 && data[0] == 'n' && data[1] == 'u' && data[2] == 'l' && data[3] == 'l' {
		o.Present = false
		return nil
	}

	if err := json.Unmarshal(data, &o.Value); err != nil {
		return err
	}

	o.Present = true
	return nil
}

// MarshalJSON implements json.Marshaler.
//
// If the value is absent, it returns "null".
//
// If present, it marshals the value.
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.Present {
		return []byte("null"), nil
	}

	return json.Marshal(o.Value)
}

// Or returns the value if present; otherwise, returns the provided default value.
//
// Example:
//
//	opt1 := optional.Some(42)
//	val1 := opt.Or(100) // val1 is 42
//
//	opt2 := optional.None[int]()
//	val2 := opt2.Or(100) // val2 is 100
func (o Optional[T]) Or(defaultVal T) T {
	if o.Present {
		return o.Value
	}
	return defaultVal
}
