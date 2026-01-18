package pieces

import (
	"encoding/json"
)

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Optional utility type
// -----------------------------------------------------------------------------

// Optional represents a value that can be null or not present in JSON
type Optional[T any] struct {
	Present bool // Whether the value is present or not
	Value   T    // The actual value
}

// UnmarshalJSON implements json.Unmarshaler for Optional[T]
func (n *Optional[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Present = false
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	n.Value = value
	n.Present = true
	return nil
}

// MarshalJSON implements json.Marshaler for Optional[T]
func (n Optional[T]) MarshalJSON() ([]byte, error) {
	if !n.Present {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}
