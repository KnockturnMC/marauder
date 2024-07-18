package utils

import (
	"fmt"
	"io"
	"os"
)

// OrElse resolves a pointer to a type T to the value at the pointer or a default
// value if the pointer is a nilpointer.
func OrElse[T any](nillable *T, defaultVal T) T {
	if nillable == nil {
		return defaultVal
	}

	return *nillable
}

// CopyFile copies the file from the passed input path to the output path.
func CopyFile(input, output string) error {
	inStream, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}

	defer func() { _ = inStream.Close() }()

	outStream, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}

	defer func() { _ = outStream.Close() }()

	if _, err := io.Copy(outStream, inStream); err != nil {
		return fmt.Errorf("failed to copy file streams: %w", err)
	}

	return nil
}
