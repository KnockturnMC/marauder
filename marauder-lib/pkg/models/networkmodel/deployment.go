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

	// RequiresRestart defines if the missmatch needs a restart to be resolved or if an upgrade
	// can be done on the running server.
	RequiresRestart bool `json:"requiresRestart"`

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

// ArtefactToInstall computes the artefact to install to resolve the missmatch.
func (a ArtefactMissmatch) ArtefactToInstall() *ArtefactVersionMissmatchArtefactInfo {
	if a.Update != nil {
		return &a.Update.Target
	} else if a.Install != nil {
		return &a.Install.Target
	}

	return nil
}

// ArtefactToUninstall computes the artefact to install to resolve the missmatch.
func (a ArtefactMissmatch) ArtefactToUninstall() *ArtefactVersionMissmatchArtefactInfo {
	if a.Update != nil {
		return &a.Update.Is
	} else if a.Uninstall != nil {
		return &a.Uninstall.Is
	}

	return nil
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
	// If the artefact UUID is nil, the update server state request implies a deletion of the state.
	ArtefactUUID *uuid.UUID `json:"artefactUuid,omitempty"`
}

// CheckFilled returns an err conveying if the request is filled with non-default values.
// insertable defines if the update server state request may be inserted, specifically if an artefact uuid is present.
func (r UpdateServerStateRequest) CheckFilled(insertable bool) error {
	if r.ArtefactIdentifier == "" {
		return fmt.Errorf("artefact identifier empty: %w", ErrMalformedModel)
	}

	if insertable && r.ArtefactUUID == nil {
		return fmt.Errorf("artefact uuid empty for inserting request: %w", ErrMalformedModel)
	}

	return nil
}
