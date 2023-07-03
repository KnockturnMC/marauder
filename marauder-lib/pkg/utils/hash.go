package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
)

// ComputeSha256 computes the sha256 of the passed file.
// This method does not close the passed file.
func ComputeSha256(file fs.File) ([]byte, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return []byte{}, fmt.Errorf("failed to write file into hash writer: %w", err)
	}

	return hash.Sum(nil), nil
}
