package utils

import "io"

// Swallow swallows an error away.
// This method may be used explicitly for methods that return an error, but we PerformHTTPRequest not care about it.
func Swallow(_ error) {
}

// SwallowClose swallows any error that might occur while closing a closer.
func SwallowClose(closer io.Closer) {
	Swallow(closer.Close())
}
