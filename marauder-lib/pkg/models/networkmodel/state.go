package networkmodel

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// The ServerStateType represents a specific type of ServerArtefactStateModel.
type ServerStateType string

const (
	// The TARGET type represents a state that has yet to be archived but is targeted.
	TARGET ServerStateType = "TARGET"

	// The IS state represents the current state of a server.
	//nolint:varnamelen
	IS = "IS"

	// The HISTORY state type represents a state that is no longer maintained by the server.
	HISTORY = "HISTORY"
)

// ErrUnknownServerState is returned when trying to create a server state with an unknown state type.
var ErrUnknownServerState = errors.New("cannot create server state for unknown server state")

// KnownServerStateType computes if the passed state is known by marauder.
func KnownServerStateType(state ServerStateType) bool {
	switch state {
	case IS:
		fallthrough
	case TARGET:
		fallthrough
	case HISTORY:
		return true
	}

	return false
}

// The ServerArtefactStateModel represents a servers specific relationship with an artefact.
// The state struct itself does not define if this is a target, is or historic state.
type ServerArtefactStateModel struct {
	// The UUID of the state, serving as a unique key for the entry.
	UUID uuid.UUID `db:"uuid" json:"uuid"`

	// The Server reference based on the servers uuid.
	Server uuid.UUID `db:"server" json:"server"`

	// The ArtefactIdentifier defines the shared identifier of the artefact type.
	ArtefactIdentifier string `db:"artefact_identifier" json:"artefactIdentifier"`

	// ArtefactUUID provides the uuid reference to the artefact this state belongs to.
	ArtefactUUID uuid.UUID `db:"artefact_uuid" json:"artefactUuid"`

	// The DefinitionDate represents the time this state was published to the controller instance.
	DefinitionDate time.Time `db:"definition_date" json:"definitionDate"`

	// The Type enum represents the type of the state.
	// For more information, see ServerStateType and its respective values.
	Type ServerStateType `db:"type" json:"type"`
}
