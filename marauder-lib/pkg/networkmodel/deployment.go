package networkmodel

import "github.com/google/uuid"

// A VersionDiff holds a difference between the version of the same artefact identifier.
// E.g. the artefact spellcore might be on version 1.0 while it targets version 2.0.
type VersionDiff struct {
	// The ArtefactIdentifier holds the shared identifier between the two versions.
	ArtefactIdentifier string

	// The IsArtefact holds the uuid of the current artefact deployed.
	IsArtefact uuid.UUID

	// The IsVersion defines the version of the current artefact.
	IsVersion string

	// The TargetArtefact holds the uuid of the targeted artefact to be deployed.
	TargetArtefact uuid.UUID

	// The TargetVersion defines the version of the artefact to deploy..
	TargetVersion string
}
