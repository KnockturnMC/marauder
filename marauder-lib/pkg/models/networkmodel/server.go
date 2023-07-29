package networkmodel

import (
	"github.com/google/uuid"
)

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
	OperatorRef ServerOperator `db:"-" json:"operator"`

	// OperatorIdentifier defines the identifier of the operator
	OperatorIdentifier string `db:"operator" json:"-"`

	// The Memory the server should allocate, defined in megabytes.
	Memory int64 `db:"memory" json:"memory"`

	// CPU represents the amount of cpu load the server may use.
	CPU float64 `db:"cpu" json:"cpu"`

	// Port defines the port of the server that it is running on and that should be exposed to the networks.
	Port int `db:"port" json:"port"`

	// Image defines the docker image the server should be spun up with.
	Image string `db:"image" json:"image"`

	// The Networks struct holds all networks the server model is part of.
	Networks []ServerNetwork `json:"networks"`
}

// ServerOperator represents an operator of a single node that the server is hosted on.
type ServerOperator struct {
	// Identifier is a string based unique identifier of an operator.
	Identifier string `json:"identifier"`

	// The Host represents the host url on which the operator can be found
	Host string `json:"host"`

	// Port represents the port the operator can be reached under on the Host.
	Port int `json:"port"`
}

// The ServerNetwork configuration defines what docker networks a server instance should be connected to.
type ServerNetwork struct {
	// The UUID of the network, uniquely identifying it across all existing network configurations.
	UUID uuid.UUID `db:"uuid" json:"uuid"`

	// The ServerUUID holds the uuid of the server this network belongs to.
	ServerUUID uuid.UUID `db:"server" json:"server"`

	// NetworkName holds the name of the external docker network.
	NetworkName string `db:"network_name" json:"networkName"`

	// The IPV4Address holds the potential ipv4 address of the container in the network.
	// If no address is specified, the string is empty.
	IPV4Address string `db:"ipv4_address" json:"ipv4Address"`
}
