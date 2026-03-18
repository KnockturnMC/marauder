package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	inStream, err := os.Open(filepath.Clean(input))
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}

	defer func() { _ = inStream.Close() }()

	outStream, err := os.Create(filepath.Clean(output))
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}

	defer func() { _ = outStream.Close() }()

	if _, err := io.Copy(outStream, inStream); err != nil {
		return fmt.Errorf("failed to copy file streams: %w", err)
	}

	return nil
}

func IntToByteSlice(i int32) []byte {
	return []byte{
		byte(i >> 24 & 0xFF),
		byte(i >> 16 & 0xFF),
		byte(i >> 8 & 0xFF),
		byte(i & 0xFF),
	}
}

func ByteSliceToInt(reader io.Reader) (int32, error) {
	bytes := []byte{0, 0, 0, 0}
	if _, err := reader.Read(bytes); err != nil {
		return 0, fmt.Errorf("failed to read: %w", err)
	}

	return int32(bytes[0])<<24 | int32(bytes[1])<<16 | int32(bytes[2])<<8 | int32(bytes[3])<<0, nil
}
