package builder

import (
    "fmt"
    "gitea.knockturnmc.com/marauder/lib/pkg/artefact"
    "io/fs"
)

// CreateArtefactTarball creates a new tar ball given a manifest at the specified target path.
// The method takes a rootFs file system in which it resolves the ci globs.
func CreateArtefactTarball(rootFs fs.FS, targetPath string, manifest artefact.Manifest) error {
    resolvedManifest, err := manifest.ResolveTemplates()
    if err != nil {
        return fmt.Errorf("failed to resolve manifest templates: %w", err)
    }
}
