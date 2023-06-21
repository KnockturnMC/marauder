package builder

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/goreleaser/fileglob"

	"gitea.knockturnmc.com/marauder/lib/pkg/artefact"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
)

const FileParentDirectory = "files/"

// CreateArtefactTarball creates a new tar ball given a manifest at the specified target path.
// The method takes a rootFs file system in which it resolves the ci globs.
// The target path is relative to the current working directory.
func CreateArtefactTarball(rootFs fs.FS, targetPath string, manifest artefact.Manifest) error {
	resolvedManifest, err := manifest.ResolveTemplates()
	if err != nil {
		return fmt.Errorf("failed to resolve manifest templates: %w", err)
	}

	tarFileHandle, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target archive %s: %w", targetPath, err)
	}

	tarballWriter := utils.NewFriendlyTarballWriterGZ(tarFileHandle)
	defer utils.SwallowClose(tarballWriter)

	// Include manifest in tarball
	serialisedManifest, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialise manifest: %w", err)
	}

	if err := tarballWriter.Write(serialisedManifest, tar.Header{
		Name: "manifest.json",
		Mode: 0o777,
	}); err != nil {
		return fmt.Errorf("failed to write manifest to tarball: %w", err)
	}

	// Include files specified in tarball.
	globCache := utils.NewShortestGlobPathCache()

	err = IncludeArtefactFiles(rootFs, resolvedManifest, globCache, tarballWriter)
	if err != nil {
		return fmt.Errorf("failed to include artefact files in tarball: %w", err)
	}

	return nil
}

// IncludeArtefactFiles is responsible for finding all files defined by the artefact manifest and writing it to the passed
// tarball writer.
func IncludeArtefactFiles(
	rootFs fs.FS,
	resolvedManifest artefact.Manifest,
	globCache *utils.ShortestGlobPathCache,
	tarballWriter utils.FriendlyTarballWriter,
) error {
	// Add files defined in manifest
	for _, file := range resolvedManifest.Files {
		matches, err := fileglob.Glob(file.CISourceGlob, fileglob.WithFs(rootFs))
		if err != nil {
			return fmt.Errorf("failed to glob manifest defined file %s: %w", file.CISourceGlob, err)
		}

		for _, match := range matches {
			shortestMatch, err := globCache.FindShortestMatch(file.CISourceGlob, match)
			if err != nil {
				return fmt.Errorf("failed to compute shortest path for file %s under glob %s: %w", match, file.CISourceGlob, err)
			}

			relativePath, err := filepath.Rel(shortestMatch, match)
			if err != nil {
				return fmt.Errorf("failed to compute relative path of %s to glob %s: %w", match, shortestMatch, err)
			}

			pathInTarball := filepath.Join(file.Target, relativePath)
			if err := tarballWriter.Add(rootFs, match, FileParentDirectory+pathInTarball); err != nil {
				return fmt.Errorf("failed to add file %s to tarball: %w", matches, err)
			}
		}
	}

	return nil
}
