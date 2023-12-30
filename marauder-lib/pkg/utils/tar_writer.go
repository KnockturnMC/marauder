package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/samber/mo"
)

// ClosableWriter is a utility interface for writers that may be closed.
type ClosableWriter interface {
	io.Writer
	io.Closer
}

// WriteResult represents a file written to the tarball.
type WriteResult struct {
	// PathInRootFS represents the path the file was read from
	PathInRootFS string
	// PathInTarball represents the path of the file in the tarball.
	// The optional may be empty in case the write result was not actually written to the tarball.
	PathInTarball mo.Option[string]
}

// The FriendlyTarballWriter interface represents a writer to a tarfile.
//
//go:generate mockery --name FriendlyTarballWriter
type FriendlyTarballWriter interface {
	io.Closer

	// Add writes a file from the root fs located at the filePathInFS to the tar ball at the filePathInTarball path.
	Add(rootFs fs.FS, filePathInFS string, filePathInTarball string) ([]WriteResult, error)

	// AddFile writes a file from the root fs located at the filePathInFS to the tar ball at the filePathInTarball path.
	AddFile(rootFs fs.FS, filePathInFS string, filePathInTarball string) (WriteResult, error)

	// Write writes the specific file content and the passed header directly to the tarball.
	Write(fileContent []byte, header tar.Header) error

	// AddFolder writes a whole from the root fs located at the filePathInFS to the tar ball at the filePathInTarball path.
	AddFolder(rootFs fs.FS, folderPathInFS string, folderPathInTarball string) ([]WriteResult, error)

	// WithFilter mutates this tarball writer to only write a file found to the tarball if the filter yields true for it.
	// Any existing filters are overwritten.
	WithFilter(filter func(pathInFS string, pathInTarball string) bool) FriendlyTarballWriter
}

// The FriendlyTarballWriterImpl struct acts as a utility for creating a tarball and implements the friendly tarball writer interface.
type FriendlyTarballWriterImpl struct {
	writerChain   []ClosableWriter
	tarballWriter *tar.Writer
	filter        func(string, string) bool
}

// NewFriendlyTarballWriterGZ constructs a new FriendlyTarballWriter that writes to the passed writer.
func NewFriendlyTarballWriterGZ(writer io.Writer, level int) (*FriendlyTarballWriterImpl, error) {
	gzipWriter, err := gzip.NewWriterLevel(writer, level)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer for level %d: %w", level, err)
	}

	tarballWriter := tar.NewWriter(gzipWriter)

	return &FriendlyTarballWriterImpl{
		writerChain:   []ClosableWriter{gzipWriter},
		tarballWriter: tarballWriter,
		filter: func(s string, s2 string) bool {
			return true
		},
	}, nil
}

// Close closes the friendly tarball writer and all owned writers of it.
func (f *FriendlyTarballWriterImpl) Close() error {
	var lastErr error
	if err := f.tarballWriter.Close(); err != nil {
		lastErr = err
	}

	for _, writer := range f.writerChain {
		if err := writer.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Write writes the specific file content and the passed header directly to the tarball.
func (f *FriendlyTarballWriterImpl) Write(fileContent []byte, header tar.Header) error {
	// Explicitly update the size of the file in the header
	header.Size = int64(len(fileContent))

	if err := f.tarballWriter.WriteHeader(&header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := f.tarballWriter.Write(fileContent); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	return nil
}

// Add writes a file from the root fs located at the filePathInFS to the tar ball at the filePathInTarball path.
func (f *FriendlyTarballWriterImpl) Add(rootFs fs.FS, filePathInFS string, filePathInTarball string) ([]WriteResult, error) {
	fileHandle, err := rootFs.Open(filePathInFS)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePathInFS, err)
	}

	defer SwallowClose(fileHandle)

	stat, err := fileHandle.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve state from file %s: %w", filePathInFS, err)
	}

	if stat.IsDir() {
		return f.AddFolder(rootFs, filePathInFS, filePathInTarball)
	}

	file, err := f.AddFile(rootFs, filePathInFS, filePathInTarball)
	if err != nil {
		return nil, fmt.Errorf("failed to add single file: %w", err)
	}

	return []WriteResult{file}, nil
}

// AddFile writes a file from the passed file system found at the filePathInFS to the tarball.
func (f *FriendlyTarballWriterImpl) AddFile(rootFs fs.FS, filePathInFS string, filePathInTarball string) (WriteResult, error) {
	open, err := rootFs.Open(filePathInFS)
	if err != nil {
		return WriteResult{}, fmt.Errorf("failed to write file to tarball, cannot open %s: %w", filePathInFS, err)
	}

	defer SwallowClose(open)

	stat, err := open.Stat()
	if err != nil {
		return WriteResult{}, fmt.Errorf("failed to read stat of file %s: %w", filePathInFS, err)
	}

	// Check the filter, do not add file if filter yields false.
	// Yield back no error but a write result without a filepath in the tarball.
	if !f.filter(filePathInFS, filePathInTarball) {
		return WriteResult{PathInRootFS: filePathInFS}, nil
	}

	header, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return WriteResult{}, fmt.Errorf("failed to construct tar header for %s: %w", filePathInFS, err)
	}

	header.Name = filePathInTarball
	if err := f.tarballWriter.WriteHeader(header); err != nil {
		return WriteResult{}, fmt.Errorf("failed to write tarball header for %s: %w", filePathInFS, err)
	}

	if _, err := io.Copy(f.tarballWriter, open); err != nil {
		return WriteResult{}, fmt.Errorf("failed to write file %s to tarball: %w", filePathInFS, err)
	}

	return WriteResult{
		PathInRootFS:  filePathInFS,
		PathInTarball: mo.Some(filePathInTarball),
	}, nil
}

// AddFolder writes the entire folder found in the fs into the tarball.
func (f *FriendlyTarballWriterImpl) AddFolder(rootFs fs.FS, folderPathInFS string, folderPathInTarball string) ([]WriteResult, error) {
	results := make([]WriteResult, 0)
	if err := fs.WalkDir(rootFs, folderPathInFS, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk folder %s: %w", path, err)
		}

		if d.IsDir() { // We PerformHTTPRequest not need to read folders
			return nil
		}

		nameInTarball := strings.Replace(path, folderPathInFS, folderPathInTarball, 1)
		file, err := f.AddFile(rootFs, path, nameInTarball)
		if err != nil {
			return fmt.Errorf("failed to add file %s to tarball: %w", path, err)
		}

		results = append(results, file)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk folder %s in fs: %w", folderPathInFS, err)
	}

	return results, nil
}

// WithFilter updates the local filter var in the friendly tarball writer impl.
func (f *FriendlyTarballWriterImpl) WithFilter(filter func(pathInFS string, pathInTarball string) bool) FriendlyTarballWriter {
	f.filter = filter
	return f
}
