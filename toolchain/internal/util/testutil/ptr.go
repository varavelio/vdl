package testutil

// Pointer creates a pointer to the given value.
func Pointer[T any](v T) *T {
	return &v
}
