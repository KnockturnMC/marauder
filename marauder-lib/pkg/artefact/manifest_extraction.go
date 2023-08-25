package artefact

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"gitea.knockturnmc.com/marauder/lib/pkg"
	artefactlib "gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"
)

// ReadManifestFromTarball reads the manifest from the passed reader.
func ReadManifestFromTarball(reader io.Reader) (*artefactlib.Manifest, error) {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	defer func() { _ = gzipReader.Close() }()

	tarballReader := tar.NewReader(gzipReader)
	var manifestPtr *artefactlib.Manifest
	for {
		nextHeader, err := tarballReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("failed to read next tarball header: %w", err)
		}

		if nextHeader.Name != pkg.ManifestFileName {
			continue
		}

		manifestAsBytes, err := io.ReadAll(tarballReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest bytes from tarball: %w", err)
		}

		manifest := artefactlib.Manifest{}
		if err := json.Unmarshal(manifestAsBytes, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse manifest from json: %w", err)
		}

		manifestPtr = &manifest

		break
	}

	return manifestPtr, nil
}
