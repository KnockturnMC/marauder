package server

import "gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"

type Manager interface {
	// Stop shuts don the server passed to the manager.
	// IF the
	Stop(server networkmodel.ServerModel) error

	// Start starts the server model
	Start(server networkmodel.ServerModel) error

	// UpdateDeployments updates all deployments currently defined on the server.
	UpdateDeployments(server networkmodel.ServerModel) error
}
