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
	"strings"

	"gitea.knockturnmc.com/marauder/lib/pkg"

	"github.com/goreleaser/fileglob"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/filemodel"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
)

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
	outputManifest := resolvedManifest

	filteredTarballWriterMap := make(map[string]struct{})
	filteredTarballWriter := tarballWriter.WithFilter(func(_ string, pathInTarball string) bool {
		_, ok := filteredTarballWriterMap[pathInTarball]
		if ok {
			return false
		}

		filteredTarballWriterMap[pathInTarball] = struct{}{}
		return true
	})

	// Add files defined in manifest
	for idx, file := range resolvedManifest.Files {
		matches, err := fileglob.Glob(file.CISourceGlob, fileglob.WithFs(rootFs))
		if err != nil {
			return filemodel.Manifest{}, fmt.Errorf("failed to glob manifest defined file %s: %w", file.CISourceGlob, err)
		}

		// Honour restriction if configured
		if err := file.Restrictions.ValidateMatchAmount(len(matches)); err != nil {
			return filemodel.Manifest{}, fmt.Errorf("failed file restriction for %s: %w", file.CISourceGlob, err)
		}

		updatedFileReference := file
		for _, match := range matches {
			err := includeMatchInTarball(rootFs, globCache, filteredTarballWriter, &updatedFileReference, match)
			if err != nil {
				return outputManifest, err
			}
		}

		outputManifest.Files[idx] = updatedFileReference // Update file reference after hash inclusion
	}

	return outputManifest, nil
}

// includeMatchInTarball includes a single matched file in the rootFs in the tarball writer and the manifest.
func includeMatchInTarball(
	rootFs fs.FS,
	globCache *utils.ShortestGlobPathCache,
	tarballWriter utils.FriendlyTarballWriter,
	file *filemodel.FileReference,
	match string,
) error {
	relativePath, err := computeRelativePath(globCache, file, match)
	if err != nil {
		return err
	}

	pathInTarball := filepath.Join(file.Target, relativePath)
	addedFiles, err := tarballWriter.Add(rootFs, match, pkg.FileParentDirectoryInArtefact+pathInTarball)
	if err != nil {
		return fmt.Errorf("failed to add file %s to tarball: %w", match, err)
	}

	if file.MatchedFiles == nil {
		file.MatchedFiles = make(map[string]string)
	}

	// Write matched files to manifest file
	for _, addedFile := range addedFiles {
		addedFileInTarball, ok := addedFile.PathInTarball.Get()
		if !ok {
			continue
		}

		hash, err := utils.ComputeSha256ForFile(rootFs, addedFile.PathInRootFS)
		if err != nil {
			return fmt.Errorf("failed to compute hash for included file %s: %w", addedFile.PathInRootFS, err)
		}

		file.MatchedFiles[addedFileInTarball] = hex.EncodeToString(hash)
	}

	return nil
}

// computeRelativePath computes the relative path of the specific match to the glob of the file that matched it.
func computeRelativePath(
	globCache *utils.ShortestGlobPathCache,
	file *filemodel.FileReference,
	match string,
) (string, error) {
	shortestMatch, err := globCache.FindShortestMatch(file.CISourceGlob, match)
	if err != nil {
		return "", fmt.Errorf("failed to compute shortest path for file %s under glob %s: %w", match, file.CISourceGlob, err)
	}

	// Respect override for shortest match / source root.
	if file.CISourceRoot != nil {
		shortestMatch = *file.CISourceRoot
	}

	relativePath, err := filepath.Rel(shortestMatch, match)
	if err != nil {
		return "", fmt.Errorf("failed to compute relative path of %s to glob %s: %w", match, shortestMatch, err)
	}

	// We matched the exact file, the exact shortest match.
	// And the target on disk ends with a slash, indicating it is not a file.
	if shortestMatch == match && strings.HasSuffix(file.Target, "/") {
		relativePath = filepath.Base(match)
	}

	return relativePath, nil
}
