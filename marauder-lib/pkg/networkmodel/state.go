package networkmodel

import (
	"time"

	"github.com/google/uuid"
)

// The ServerStateType represents a specific type of ServerArtefactStateModel.
type ServerStateType string

const (
	// The TARGET type represents a state that has yet to be archived but is targeted.
	TARGET ServerStateType = "TARGET"

	// The IS state represents the current state of a server.
	IS = "IS"

	// The HISTORY state type represents a state that is no longer maintained by the server.
	HISTORY = "HISTORY"
)

// The ServerArtefactStateModel represents a servers specific relationship with an artefact.
// The state struct itself does not define if this is a target, is or historic state.
type ServerArtefactStateModel struct {
	// The UUID of the state, serving as a unique key for the entry.
	UUID uuid.UUID `db:"uuid"`

	// The Server reference based on the servers uuid.
	Server uuid.UUID `db:"server"`

	// Artefact provides the uuid reference to the artefact this state belongs to.
	Artefact uuid.UUID `db:"artefact"`

	// The DefinitionDate represents the time this state was published to the controller instance.
	DefinitionDate time.Time `db:"definition_date"`

	// The Type enum represents the type of the state.
	// For more information, see ServerStateType and its respective values.
	Type ServerStateType `db:"type"`
}
