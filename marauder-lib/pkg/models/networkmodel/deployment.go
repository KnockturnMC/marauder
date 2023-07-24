package networkmodel

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// BodyModel represents a struct that is parsed from a json body.
type BodyModel interface {
	// CheckFilled returns an err conveying if the request is filled with non-default values.
	CheckFilled() error
}

// ErrMalformedModel is returned if a model is found to be in any way malformed.
var ErrMalformedModel = errors.New("malformed model")

// A VersionDiff holds a difference between the version of the same artefact identifier.
// E.g. the artefact spellcore might be on version 1.0 while it targets version 2.0.
type VersionDiff struct {
	// The ArtefactIdentifier holds the shared identifier between the two versions.
	ArtefactIdentifier string `db:"artefact_identifier" json:"artefactIdentifier"`

	// The IsArtefact holds the uuid of the current artefact deployed.
	IsArtefact uuid.UUID `db:"is_artefact" json:"isArtefact"`

	// The IsVersion defines the version of the current artefact.
	IsVersion string `db:"is_version" json:"isVersion"`

	// The TargetArtefact holds the uuid of the targeted artefact to be deployed.
	TargetArtefact uuid.UUID `db:"target_artefact" json:"targetArtefact"`

	// The TargetVersion defines the version of the artefact to deploy.
	TargetVersion string `db:"target_version" json:"targetVersion"`
}

// The UpdateServerStateRequest is pushed to the deployment endpoint as a body to trigger an update to a new server state.
type UpdateServerStateRequest struct {
	// The ArtefactIdentifier defines the shared identifier of the artefact type.
	ArtefactIdentifier string `json:"artefactIdentifier"`

	// ArtefactUUID provides the uuid reference to the artefact this state belongs to.
	ArtefactUUID uuid.UUID `json:"artefactUuid"`
}

// CheckFilled returns an err conveying if the request is filled with non-default values.
func (r UpdateServerStateRequest) CheckFilled() error {
	if r.ArtefactIdentifier == "" {
		return fmt.Errorf("artefact identifier empty: %w", ErrMalformedModel)
	}
	if r.ArtefactUUID.Version() == 0 {
		return fmt.Errorf("artefact uuid empty: %w", ErrMalformedModel)
	}

	return nil
}
