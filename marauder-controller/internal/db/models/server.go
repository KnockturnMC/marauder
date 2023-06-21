package models

import "github.com/google/uuid"

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
}

// The ServerDependencyModel struct represents a servers dependency on a given artefact by its identifier.
// This is not versioned as it does neither represent the *is*  and *target* state, it simply configures the server to have the artefact installed.
type ServerDependencyModel struct {
	// The Server uuid references the server this dependency configures.
	Server uuid.UUID `db:"server"`

	// The ArtefactIdentifier represents the identifier of the artefact that the server requires to operate.
	ArtefactIdentifier string `db:"artefact_identifier"`
}

// The ServerDockerConfigurationModel represents a configuration for a specific server, detailing its docker container configuration.
type ServerDockerConfigurationModel struct {
	// The Server uuid references the server this docker configuration belongs to.
	Server uuid.UUID `db:"server"`

	// Image defines the docker image the server should be spun up with.
	Image string `db:"image"`

	// CPUs defines how many cpus the server instance should be allocated with when running.
	CPUs float64 `db:"cpus"`

	// The Memory the server should allocate, defined in megabytes.
	Memory int64 `db:"memory"`
}
