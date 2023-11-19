package utils

// OrElse resolves a pointer to a type T to the value at the pointer or a default
// value if the pointer is a nilpointer.
func OrElse[T any](nillable *T, defaultVal T) T {
	if nillable == nil {
		return defaultVal
	}

	return *nillable
}
