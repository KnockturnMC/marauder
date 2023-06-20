package artefact

import (
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
)

// The Manifest type defines an artefact's manifest managed by marauder.
type Manifest struct {
	// The unique, marauder wide Identifier of the artefact, usually the name of the plugin the artefact is created for.
	// An example would be `spellcore`.
	Identifier string `json:"identifier"`

	// The Version of the artefact, following schematic versioning.
	Version string `json:"version"`

	// The Files included in this artefact, not flattened.
	// The file reference may hence include specific files or references to whole folders in the artefact.
	Files []FileReference `json:"files"`
}

// ResolveTemplates constructs a new manifest that has its go templates resolved.
// This uses golang templating engine to resolve the templates that may exist in the manifest.
func (m Manifest) ResolveTemplates() (Manifest, error) {
	mappedFiles := make([]FileReference, 0, len(m.Files))
	for i, file := range m.Files {
		updated, err := file.ResolveTemplates(m)
		if err != nil {
			return Manifest{}, fmt.Errorf("failed to resolve file ref %d: %w", i, err)
		}

		mappedFiles = append(mappedFiles, updated)
	}

	return Manifest{
		Identifier: m.Identifier,
		Version:    m.Version,
		Files:      mappedFiles,
	}, nil
}

// FileReference defines a specific configuration of an artefacts file as defined by its manifest.
type FileReference struct {
	// The target path as found in the artefacts file collection as well as its final location on the server.
	Target string `json:"target"`

	// A string representation of a glob that identifies the files during the ci build of the project that produces the artefact.
	CISourceGlob string `json:"ciSourceGlob"`
}

// ResolveTemplates constructs a new file reference that has its go templates resolved.
// This uses golang templating engine to resolve the templates that may exist in the file reference.
func (f FileReference) ResolveTemplates(data Manifest) (FileReference, error) {
	resolvedTarget, err := utils.ExecuteStringTemplateToString(f.Target, data)
	if err != nil {
		return FileReference{}, fmt.Errorf("failed to resolve `target`: %w", err)
	}

	resolvedCISourceGlob, err := utils.ExecuteStringTemplateToString(f.CISourceGlob, data)
	if err != nil {
		return FileReference{}, fmt.Errorf("failed to resolve `ciSourceGlob`: %w", err)
	}

	return FileReference{
		Target:       resolvedTarget,
		CISourceGlob: resolvedCISourceGlob,
	}, nil
}
