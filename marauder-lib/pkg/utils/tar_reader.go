package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// A FriendlyTarballReader is a tarball reader that wraps an existing io.Reader and optionally closes it.
type FriendlyTarballReader struct {
	*tar.Reader

	gzipReader *gzip.Reader
	ioReader   io.Reader
}

// NewFriendlyTarballReaderFromPath constructs a new reader from the file at the path on the disk.
func NewFriendlyTarballReaderFromPath(path string) (*FriendlyTarballReader, error) {
	ioReader, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return NewFriendlyTarballReaderFromReader(ioReader)
}

// NewFriendlyTarballReaderFromFS constructs a new reader from the file at the path on the fs.
func NewFriendlyTarballReaderFromFS(fs fs.FS, path string) (*FriendlyTarballReader, error) {
	ioReader, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return NewFriendlyTarballReaderFromReader(ioReader)
}

// NewFriendlyTarballReaderFromReader constructs a new reader from the passed reader.
func NewFriendlyTarballReaderFromReader(reader io.Reader) (*FriendlyTarballReader, error) {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzipped reader: %w", err)
	}

	return &FriendlyTarballReader{
		Reader:     tar.NewReader(gzipReader),
		gzipReader: gzipReader,
		ioReader:   reader,
	}, nil
}

// Close closes the underlying gzip reader and potentially the io reader if defined.
func (f *FriendlyTarballReader) Close(closeIoReader bool) error {
	defer func() {
		if !closeIoReader {
			return
		}

		closer, ok := f.ioReader.(io.Closer)
		if ok {
			_ = closer.Close()
		}
	}()

	if err := f.gzipReader.Close(); err != nil {
		return fmt.Errorf("failed to close gzip reader: %w", err)
	}

	return nil
}
