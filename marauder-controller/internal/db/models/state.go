package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrServerStateTypeUnknown is returned if a value of ServerStateType is used that is unknown by marauder.
var ErrServerStateTypeUnknown = errors.New("unknown ServerStateType")

// The ServerStateType represents a specific type of ServerArtefactState.
type ServerStateType uint

// ToSQL converts a server state type into the SQL enum representation.
func (s ServerStateType) ToSQL() (string, error) {
	switch s {
	case TARGET:
		return "TARGET", nil
	case IS:
		return "IS", nil
	case HISTORY:
		return "HISTORY", nil
	}

	return "", fmt.Errorf("passed server state type is unknown %d: %w", s, ErrServerStateTypeUnknown)
}

const (
	// The TARGET type represents a state that has yet to be archived but is targeted.
	TARGET ServerStateType = iota

	// The IS state represents the current state of a server.
	IS

	// The HISTORY state type represents a state that is no longer maintained by the server.
	HISTORY
)

// The ServerArtefactState represents a servers specific relationship with an artefact.
// The state struct itself does not define if this is a target, is or historic state.
type ServerArtefactState struct {
	UUID uuid.UUID `db:"uuid"`

	// The Server reference based on the servers uuid.
	Server uuid.UUID `db:"server"`

	Artefact uuid.UUID `db:"artefact"`

	// The DefinitionDate represents the time this state was published to the controller instance.
	DefinitionDate time.Time

	// The Type enum represents the type of the state.
	// For more information, see ServerStateType and its respective values.
	Type ServerStateType
}
