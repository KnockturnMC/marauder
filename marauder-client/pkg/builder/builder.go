package builder

import (
	"archive/tar"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"gitea.knockturnmc.com/marauder/lib/pkg"

	"github.com/goreleaser/fileglob"

	"gitea.knockturnmc.com/marauder/lib/pkg/filemodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
)

const FileParentDirectory = "files/"

// CreateArtefactTarball creates a new tar ball given a manifest at the specified target path.
// The method takes a rootFs file system in which it resolves the ci globs.
// The target path is relative to the current working directory.
func CreateArtefactTarball(rootFs fs.FS, manifest filemodel.Manifest, writer io.Writer) error {
	tarballWriter, err := utils.NewFriendlyTarballWriterGZ(writer, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("faild to create friendly tarball writer: %w", err)
	}

	defer utils.SwallowClose(tarballWriter)

	// Include files specified in tarball.
	globCache := utils.NewShortestGlobPathCache()
	manifest, err = IncludeArtefactFiles(rootFs, manifest, globCache, tarballWriter)
	if err != nil {
		return fmt.Errorf("failed to include artefact files in tarball: %w", err)
	}

	// Include manifest in tarball
	serialisedManifest, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialise manifest: %w", err)
	}

	if err := tarballWriter.Write(serialisedManifest, tar.Header{
		Name: pkg.ManifestFileName,
		Mode: 0o777,
	}); err != nil {
		return fmt.Errorf("failed to write manifest to tarball: %w", err)
	}

	return nil
}

// IncludeArtefactFiles is responsible for finding all files defined by the artefact manifest and writing it to the passed
// tarball writer.
func IncludeArtefactFiles(
	rootFs fs.FS,
	resolvedManifest filemodel.Manifest,
	globCache *utils.ShortestGlobPathCache,
	tarballWriter utils.FriendlyTarballWriter,
) (filemodel.Manifest, error) {
	// Create Hashes map if needed.
	if resolvedManifest.Hashes == nil {
		resolvedManifest.Hashes = make(map[string]string)
	}

	// Add files defined in manifest
	for _, file := range resolvedManifest.Files {
		matches, err := fileglob.Glob(file.CISourceGlob, fileglob.WithFs(rootFs))
		if err != nil {
			return filemodel.Manifest{}, fmt.Errorf("failed to glob manifest defined file %s: %w", file.CISourceGlob, err)
		}

		for _, match := range matches {
			shortestMatch, err := globCache.FindShortestMatch(file.CISourceGlob, match)
			if err != nil {
				return filemodel.Manifest{}, fmt.Errorf("failed to compute shortest path for file %s under glob %s: %w", match, file.CISourceGlob, err)
			}

			relativePath, err := filepath.Rel(shortestMatch, match)
			if err != nil {
				return filemodel.Manifest{}, fmt.Errorf("failed to compute relative path of %s to glob %s: %w", match, shortestMatch, err)
			}

			pathInTarball := filepath.Join(file.Target, relativePath)
			addedFiles, err := tarballWriter.Add(rootFs, match, FileParentDirectory+pathInTarball)
			if err != nil {
				return filemodel.Manifest{}, fmt.Errorf("failed to add file %s to tarball: %w", matches, err)
			}

			// Write hashes
			for _, addedFile := range addedFiles {
				hashArray, err := computeHashFor(rootFs, addedFile.PathInRootFS)
				if err != nil {
					return filemodel.Manifest{}, fmt.Errorf("faild to compute hash for %s: %w", addedFile.PathInRootFS, err)
				}

				resolvedManifest.Hashes[addedFile.PathInTarball] = hex.EncodeToString(hashArray)
			}
		}
	}

	return resolvedManifest, nil
}

// computeHashFor computes a sha256 hash for the file located at the given path in the given file system.
func computeHashFor(rootFs fs.FS, path string) ([]byte, error) {
	open, err := rootFs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s for hashsum computation: %w", path, err)
	}

	defer func() { _ = open.Close() }()
	sha256, err := utils.ComputeSha256(open)
	if err != nil {
		return nil, fmt.Errorf("failed to compute sha 256 hash for file: %w", err)
	}

	return sha256, nil
}
