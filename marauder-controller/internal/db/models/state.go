package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrServerStateTypeUnknown is returned if a value of ServerStateType is used that is unknown by marauder.
var ErrServerStateTypeUnknown = errors.New("unknown ServerStateType")

// The ServerStateType represents a specific type of ServerArtefactStateModel.
type ServerStateType uint

const (
	// The TARGET type represents a state that has yet to be archived but is targeted.
	TARGET ServerStateType = iota

	// The IS state represents the current state of a server.
	IS

	// The HISTORY state type represents a state that is no longer maintained by the server.
	HISTORY
)

// Value converts a server state type to a driver applicable value.
func (s *ServerStateType) Value() (driver.Value, error) {
	switch *s {
	case HISTORY:
		return "HISTORY", nil
	case IS:
		return "IS", nil
	case TARGET:
		return "TARGET", nil
	}
	return nil, fmt.Errorf("failed to build value for sql: %w", ErrServerStateTypeUnknown)
}

// Scan scans the server state type from the provided source.
func (s *ServerStateType) Scan(src any) error {
	stringSrc, ok := src.(string)
	if !ok {
		return fmt.Errorf("failed to scan server state type, not a string: %w", ErrServerStateTypeUnknown)
	}
	switch stringSrc {
	case "HISTORY":
		*s = HISTORY
	case "TARGET":
		*s = TARGET
	case "IS":
		*s = IS
	default:
		return fmt.Errorf("failed to scan server state type %s: %w", stringSrc, ErrServerStateTypeUnknown)
	}

	return nil
}

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
