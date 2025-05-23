package artefact

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"

	"gitea.knockturnmc.com/marauder/lib/pkg"
	artefactlib "gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"
)

// ReadManifestFromTarball reads the manifest from the passed reader.
func ReadManifestFromTarball(reader io.Reader) (*artefactlib.Manifest, error) {
	tarballReader, err := utils.NewFriendlyTarballReaderFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create tarbaöö reader: %w", err)
	}

	defer func() { _ = tarballReader.Close(false) }()

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
func ReadManifestAtCurrent(tarballReader *tar.Reader) (artefactlib.Manifest, error) {
	manifestAsBytes, err := io.ReadAll(tarballReader)
	if err != nil {
		return artefactlib.Manifest{}, fmt.Errorf("failed to read manifest bytes from tarball: %w", err)
	}

	manifest := artefactlib.Manifest{}
	if err := json.Unmarshal(manifestAsBytes, &manifest); err != nil {
		return artefactlib.Manifest{}, fmt.Errorf("failed to parse manifest from json: %w", err)
	}
	return manifest, nil
}
