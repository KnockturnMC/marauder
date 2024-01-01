package artefact

import (
	"archive/tar"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Goldziher/go-utils/sliceutils"

	"gitea.knockturnmc.com/marauder/lib/pkg"
	artefactlib "gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/Goldziher/go-utils/maputils"
)

// ErrHashMissmatch is yielded if the validator finds a file that does not match its defined hash.
var ErrHashMissmatch = errors.New("hash missmatch")

// verifyArtefactManifestHashes verifies the included hashes in the artefacts manifest file
// ensuring that the hashes defined in the manifest of the artefact are correct compared to the files
// included.
func (w *WorkedBasedValidator) verifyArtefactManifestHashes(artefact *os.File) (artefactlib.Manifest, error) {
	if _, err := artefact.Seek(0, io.SeekStart); err != nil {
		return artefactlib.Manifest{}, fmt.Errorf("failed to reset artefact file ref to start: %w", err)
	}

	tarReader, err := utils.NewFriendlyTarballReaderFromReader(artefact)
	if err != nil {
		return artefactlib.Manifest{}, fmt.Errorf("failed to create tarball reader: %w", err)
	}
	defer func() { _ = tarReader.Close(false) }()

	manifest, filesIncluded, err := readTarballForValidation(tarReader.Reader)
	if err != nil {
		return artefactlib.Manifest{}, fmt.Errorf("failed to parse tarball: %w", err)
	}

	filesToHashesFromManifest := maputils.Merge(
		sliceutils.Map(
			manifest.Files,
			func(value artefactlib.FileReference, index int, slice []artefactlib.FileReference) map[string]string {
				return value.MatchedFiles
			},
		)...,
	)
	if len(filesToHashesFromManifest) != len(filesIncluded) {
		return artefactlib.Manifest{}, fmt.Errorf(
			"found %d files, expected %d: %w",
			len(filesIncluded),
			len(filesToHashesFromManifest),
			ErrUnaccountedForFile,
		)
	}

	for filePath, expectedHash := range filesToHashesFromManifest {
		foundHash, ok := filesIncluded[filePath]
		if !ok {
			return artefactlib.Manifest{}, fmt.Errorf("file %s not found in artefact but was defined: %w", filePath, ErrUnaccountedForFile)
		}

		if foundHash != expectedHash {
			return artefactlib.Manifest{}, fmt.Errorf("file %s did not match expected hash: %w", filePath, ErrHashMissmatch)
		}

		delete(filesIncluded, filePath)
	}

	if len(filesIncluded) != 0 {
		return artefactlib.Manifest{}, fmt.Errorf(
			"artefact contained %d files not defined in the manifest (%s): %w",
			len(filesIncluded),
			strings.Join(maputils.Keys(filesIncluded), ", "),
			ErrUnaccountedForFile,
		)
	}

	return manifest, nil
}

// readTarballForValidation reads in the entire tarball from the passed tarball reader and returns the manifest of the tarball
// as well as a map of file paths in the tarball.
func readTarballForValidation(tarReader *tar.Reader) (artefactlib.Manifest, map[string]string, error) {
	var manifest *artefactlib.Manifest
	includedFiles := make(map[string]string)

	// Read entire tar file, computing all included files and storing them.
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return artefactlib.Manifest{}, nil, fmt.Errorf("failed to read header from artefact tarball: %w", err)
		}

		if header.Name == pkg.ManifestFileName {
			manifestBytes, err := io.ReadAll(tarReader)
			if err != nil {
				return artefactlib.Manifest{}, nil, fmt.Errorf("failed to read byte of amnifest file from tarball: %w", err)
			}

			manifest = &artefactlib.Manifest{}
			if err := json.Unmarshal(manifestBytes, manifest); err != nil {
				return artefactlib.Manifest{}, nil, fmt.Errorf("failed to parse manifest file: %w", err)
			}
		} else {
			fileHash, err := utils.ComputeSha256(tarReader)
			if err != nil {
				return artefactlib.Manifest{}, nil, fmt.Errorf("failed to compute hash for file %s: %w", header.Name, err)
			}

			includedFiles[header.Name] = hex.EncodeToString(fileHash)
		}
	}

	// Return error if manifest is missing
	if manifest == nil {
		return artefactlib.Manifest{}, nil, ErrManifestMissing
	}

	return *manifest, includedFiles, nil
}
