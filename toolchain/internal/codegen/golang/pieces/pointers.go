package pieces

/** START FROM HERE **/

// -----------------------------------------------------------------------------
// Pointer utility functions
// -----------------------------------------------------------------------------

// Ptr returns a pointer to the value.
func Ptr[T any](value T) *T {
	return &value
}

// Val returns the value pointed to by pointer, or the zero value of T if pointer is nil.
func Val[T any](pointer *T) T {
	if pointer == nil {
		var zero T
		return zero
	}
	return *pointer
}

// Or returns the value pointed to by pointer, or defaultValue if pointer is nil.
func Or[T any](pointer *T, defaultValue T) T {
	if pointer == nil {
		return defaultValue
	}
	return *pointer
}
