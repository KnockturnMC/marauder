package models

import "time"

// ArtefactModel represents an artefact entry in the database.
type ArtefactModel struct {
	// A unique, string based Identifier of the artefact as laid out in its manifest. Examples include `spellcore` or `knockturncore`.
	Identifier string `db:"identifier"`

	// The Version of the artefact. This version follows schematic versioning rules and, in combination with the identifier, uniquely identifies
	// an artefact in the controller.
	Version string `db:"version"`

	// The UploadDate of the artefact to the controller.
	UploadDate time.Time `db:"upload_date"`

	// The StoragePath of the artefact tarball in the controllers main storage directory.
	// This might be a combination of the identifier and version, but is not guaranteed and hence is defined here.
	StoragePath string `db:"storage_path"`
}
