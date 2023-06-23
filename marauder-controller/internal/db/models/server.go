package models

import (
	"github.com/google/uuid"
)

// The ServerModel struct represents a servers configuration in the database model.
type ServerModel struct {
	// UUID is a unique identifier of the server instance across all environment and names.
	UUID uuid.UUID `db:"uuid"`

	// Environment represents what environment the server lives in. `production` indicating the production environment,
	// `integration` the integration environment etc.
	Environment string `db:"environment"`

	// Name serves as a display name of the server that can be returned to make interaction with the server easier.
	Name string `db:"name"`

	// Host represents the host the server can be found on. This may be an internal url, however does not need to be.
	// The controller simply has to be able to locate an operator based on the host defined for the server.
	Host string `db:"host"`

	// The Memory the server should allocate, defined in megabytes.
	Memory int64 `db:"memory"`

	// CPU represents the amount of cpu load the server may use.
	CPU float64 `db:"cpu"`

	// Image defines the docker image the server should be spun up with.
	Image string `db:"image"`
}
