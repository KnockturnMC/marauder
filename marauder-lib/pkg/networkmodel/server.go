package networkmodel

import "github.com/google/uuid"

// The ServerModel struct represents a servers configuration in the database model.
type ServerModel struct {
	// UUID is a unique identifier of the server instance across all environment and names.
	UUID uuid.UUID `db:"uuid" json:"uuid"`

	// Environment represents what environment the server lives in. `production` indicating the production environment,
	// `integration` the integration environment etc.
	Environment string `db:"environment" json:"environment"`

	// Name serves as a display name of the server that can be returned to make interaction with the server easier.
	Name string `db:"name" json:"name"`

	// Host represents the host the server can be found on. This may be an internal url, however does not need to be.
	// The controller simply has to be able to locate an operator based on the host defined for the server.
	Host string `db:"host" json:"host"`

	// The Memory the server should allocate, defined in megabytes.
	Memory int64 `db:"memory" json:"memory"`

	// CPU represents the amount of cpu load the server may use.
	CPU float64 `db:"cpu" json:"cpu"`

	// Image defines the docker image the server should be spun up with.
	Image string `db:"image" json:"image"`

	// The Networks struct holds all networks the server model is part of.
	Networks []ServerNetwork `json:"networks"`
}

// The ServerNetwork configuration defines what docker networks a server instance should be connected to.
type ServerNetwork struct {
	// The UUID of the network, uniquely identifying it across all existing network configurations.
	UUID uuid.UUID `db:"uuid"`

	// The ServerUUID holds the uuid of the server this network belongs to.
	ServerUUID uuid.UUID `db:"server"`

	// NetworkName holds the name of the external docker network.
	NetworkName string `db:"network_name"`

	// The IPV4Address holds the potential ipv4 address of the container in the network.
	// If no address is specified, the string is empty.
	IPV4Address string `db:"ipv4_address"`
}
