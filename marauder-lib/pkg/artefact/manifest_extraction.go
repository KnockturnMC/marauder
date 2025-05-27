package artefact

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/knockturnmc/marauder/marauder-lib/pkg"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/filemodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
)

// ReadManifestFromTarball reads the manifest from the passed reader.
func ReadManifestFromTarball(reader io.Reader) (*filemodel.Manifest, error) {
	tarballReader, err := utils.NewFriendlyTarballReaderFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create tarbaöö reader: %w", err)
	}

	defer func() { _ = tarballReader.Close(false) }()

	var manifestPtr *filemodel.Manifest
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

		manifest, err := ReadManifestAtCurrent(tarballReader.Reader)
		if err != nil {
			return nil, err
		}

		manifestPtr = &manifest

		break
	}

	return manifestPtr, nil
}

// ReadManifestAtCurrent reads the manifest file from the tar.Reader at the current header.
func ReadManifestAtCurrent(tarballReader *tar.Reader) (filemodel.Manifest, error) {
	manifestAsBytes, err := io.ReadAll(tarballReader)
	if err != nil {
		return filemodel.Manifest{}, fmt.Errorf("failed to read manifest bytes from tarball: %w", err)
	}

	manifest := filemodel.Manifest{}
	if err := json.Unmarshal(manifestAsBytes, &manifest); err != nil {
		return filemodel.Manifest{}, fmt.Errorf("failed to parse manifest from json: %w", err)
	}
	return manifest, nil
}
