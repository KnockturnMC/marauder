package models

import (
	"time"

	"github.com/google/uuid"
)

// ArtefactModel represents an artefact entry in the database.
type ArtefactModel struct {
	// The unique identifier of the artefacts record in the database.
	UUID uuid.UUID `db:"uuid"`

	// A unique, string based Identifier of the artefact as laid out in its manifest. Examples include `spellcore` or `knockturncore`.
	Identifier string `db:"identifier"`

	// The Version of the artefact. This version follows schematic versioning rules and, in combination with the identifier, uniquely identifies
	// an artefact in the controller.
	Version string `db:"version"`

	// The UploadDate of the artefact to the controller.
	UploadDate time.Time `db:"upload_date"`
}

// The ArtefactModelWithBinary struct represents a full artefact, including its tarball.
type ArtefactModelWithBinary struct {
	ArtefactModel

	// The TarballBlob holds the entire tarball of the artefact
	TarballBlob []byte `db:"tarball"`
}
