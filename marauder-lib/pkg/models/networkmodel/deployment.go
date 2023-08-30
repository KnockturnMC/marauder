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

// A ArtefactVersionMissmatch holds a difference between the version of the same artefact identifier.
// E.g. the artefact spellcore might be on version 1.0 while it targets version 2.0.
type ArtefactVersionMissmatch struct {
	// The ArtefactIdentifier holds the shared identifier between the two versions.
	ArtefactIdentifier string `json:"artefactIdentifier"`

	// The Missmatch defines the missmatch container that holds the potential missmatch.
	Missmatch ArtefactMissmatch `json:"missmatch"`
}

type ArtefactMissmatch struct {
	// Update defines that a missmatch exists because the artefact already exists on the server but needs to be updated to a new version.
	Update *ArtefactVersionMissmatchUpdate `json:"update,omitempty"`

	// Install defines that the server needs to install an artefact that is not running as of right now.
	Install *ArtefactVersionMissmatchInstall `json:"install,omitempty"`

	// Uninstall defines that the server is running an artefact that is not needed to be running.
	Uninstall *ArtefactVersionMissmatchUninstall `json:"uninstall,omitempty"`
}

// ArtefactVersionMissmatchUpdate defines that a version of an artefact is out of date and needs to be updated.
type ArtefactVersionMissmatchUpdate struct {
	// Is defines the current version the server is running of the artefact.
	Is ArtefactVersionMissmatchArtefactInfo `json:"is"`

	// Target defines the target version the server is supposed to be running of the artefact.
	Target ArtefactVersionMissmatchArtefactInfo `json:"target"`
}

// ArtefactVersionMissmatchInstall defines that a server is not yet running a specific artefact, but should.
type ArtefactVersionMissmatchInstall struct {
	// Target defines the target version the server is supposed to be running of the artefact.
	Target ArtefactVersionMissmatchArtefactInfo `json:"target"`
}

// ArtefactVersionMissmatchUninstall defines that a server is running a specific artefact, but shouldn't.
type ArtefactVersionMissmatchUninstall struct {
	// Is defines the current version the server is running of the artefact.
	Is ArtefactVersionMissmatchArtefactInfo `json:"is"`
}

type ArtefactVersionMissmatchArtefactInfo struct {
	// The Artefact holds the uuid of the artefact this info describes.
	Artefact uuid.UUID `db:"artefact" json:"artefact"`

	// The Version defines the version of the artefact this info describes.
	Version string `db:"version" json:"version"`
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
