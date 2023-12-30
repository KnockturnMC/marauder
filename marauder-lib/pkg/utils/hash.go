package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
)

// ComputeSha256 computes the sha256 of the passed file.
// This method does not close the passed file.
func ComputeSha256(file io.Reader) ([]byte, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return []byte{}, fmt.Errorf("failed to write file into hash writer: %w", err)
	}

	return hash.Sum(nil), nil
}

// ComputeSha256ForFile computes a sha256 hash for the file located at the given path in the given file system.
func ComputeSha256ForFile(rootFs fs.FS, path string) ([]byte, error) {
	open, err := rootFs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s for hashsum computation: %w", path, err)
	}

	defer func() { _ = open.Close() }()
	hash, err := ComputeSha256(open)
	if err != nil {
		return nil, fmt.Errorf("failed to compute sha 256 hash for file: %w", err)
	}

	return hash, nil
}
