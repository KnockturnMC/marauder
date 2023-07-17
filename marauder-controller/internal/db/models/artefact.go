package models

import (
	"time"

	"github.com/google/uuid"
)

// ArtefactModel represents an artefact entry in the database.
type ArtefactModel struct {
	// The unique identifier of the artefacts record in the database.
	UUID uuid.UUID `db:"uuid" json:"uuid"`

	// A unique, string based Identifier of the artefact as laid out in its manifest. Examples include `spellcore` or `knockturncore`.
	Identifier string `db:"identifier" json:"identifier"`

	// The Version of the artefact. This version follows schematic versioning rules and, in combination with the identifier, uniquely identifies
	// an artefact in the controller.
	Version string `db:"version" json:"version"`

	// The UploadDate of the artefact to the controller.
	UploadDate time.Time `db:"upload_date" json:"uploadDate"`
}

// The ArtefactModelWithBinary struct represents a full artefact, including its tarball.
type ArtefactModelWithBinary struct {
	ArtefactModel

	// The TarballBlob holds the entire tarball of the artefact
	TarballBlob []byte `db:"tarball"`

	// The Hash of the tarball this artefact represents in the format of a sha256 hash.,
	Hash []byte `db:"hash" json:"hash"`
}
