package artefact

import (
	"archive/tar"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gitea.knockturnmc.com/marauder/lib/pkg"
	artefactlib "gitea.knockturnmc.com/marauder/lib/pkg/artefact"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"github.com/Goldziher/go-utils/maputils"
)

// verifyArtefactManifestHashes verifies the included hashes in the artefacts manifest file
// ensuring that the hashes defined in the manifest of the artefact are correct compared to the files
// included.
func (w *WorkedBasedValidator) verifyArtefactManifestHashes(artefact *os.File) (artefactlib.Manifest, error) {
	if _, err := artefact.Seek(0, io.SeekStart); err != nil {
		return artefactlib.Manifest{}, fmt.Errorf("failed to reset artefact file ref to start: %w", err)
	}

	gzipReader, err := gzip.NewReader(artefact)
	if err != nil {
		return artefactlib.Manifest{}, fmt.Errorf("failed to create gzip reader for artefact: %w", err)
	}

	defer func() { _ = gzipReader.Close() }()

	tarReader := tar.NewReader(gzipReader)

	manifest, fileToHashMap, err := readTarballForValidation(tarReader)
	if err != nil {
		return artefactlib.Manifest{}, fmt.Errorf("failed to parse tarball: %w", err)
	}

	if len(manifest.Hashes) != len(fileToHashMap) {
		return artefactlib.Manifest{}, fmt.Errorf("found %d files and %d hashes: %w", len(fileToHashMap), len(manifest.Hashes), ErrHashMismatch)
	}

	for filePath, fileHash := range manifest.Hashes {
		foundHash, ok := fileToHashMap[filePath]
		if !ok {
			return artefactlib.Manifest{}, fmt.Errorf("file %s not found in artefact but had hash defined: %w", filePath, ErrHashMismatch)
		}

		delete(fileToHashMap, filePath)

		foundHashInHex := hex.EncodeToString(foundHash)
		if fileHash != foundHashInHex {
			return artefactlib.Manifest{}, fmt.Errorf("file %s should have hash %s, got %s: %w", filePath, fileHash, foundHashInHex, ErrHashMismatch)
		}
	}

	if len(fileToHashMap) != 0 {
		return artefactlib.Manifest{}, fmt.Errorf(
			"artefact contained %d files not defined in the manifest (%s): %w",
			len(fileToHashMap),
			strings.Join(maputils.Keys(fileToHashMap), ", "),
			ErrHashMismatch,
		)
	}

	return manifest, nil
}

// readTarballForValidation reads in the entire tarball from the passed tarball reader and returns the manifest of the tarball
// as well as a map of file paths in the tarball to their sha256 hashes.
func readTarballForValidation(tarReader *tar.Reader) (artefactlib.Manifest, map[string][]byte, error) {
	var manifest artefactlib.Manifest
	fileToHashMap := make(map[string][]byte)

	// Read entire tar file, computing all hashes and storing them.
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

			if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
				return artefactlib.Manifest{}, nil, fmt.Errorf("failed to parse manifest file: %w", err)
			}
		} else {
			sha256, err := utils.ComputeSha256(tarReader)
			if err != nil {
				return artefactlib.Manifest{}, nil, fmt.Errorf("failed to compute sha hash for file %s: %w", header.Name, err)
			}

			fileToHashMap[header.Name] = sha256
		}
	}

	return manifest, fileToHashMap, nil
}
