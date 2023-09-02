package access

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
)

// fillServerModelNetwork fetches the network configuration of a given server model from the database.
func fillServerModelRefs(ctx context.Context, db *sqlm.DB, model networkmodel.ServerModel) (networkmodel.ServerModel, error) {
	var err error
	if model, err = fillServerModelNetwork(ctx, db, model); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to fetch networks: %w", err)
	}

	if model, err = fillServerModelOperator(ctx, db, model); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to fetch operator: %w", err)
	}

	if model, err = fillServerModelHostPorts(ctx, db, model); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to fetch host ports: %w", err)
	}

	return model, nil
}

// fillServerModelNetworkSlice executes fillServerModelNetwork for each server in the slice.
func fillServerModelRefsSlice(ctx context.Context, db *sqlm.DB, servers []networkmodel.ServerModel) ([]networkmodel.ServerModel, error) {
	for i, el := range servers {
		fullyFetched, err := fillServerModelRefs(ctx, db, el)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch refs for %s: %w", el.UUID.String(), err)
		}
		servers[i] = fullyFetched
	}

	return servers, nil
}

// fillServerModelNetwork fetches the network configuration of a given server model from the database.
func fillServerModelNetwork(ctx context.Context, db *sqlm.DB, model networkmodel.ServerModel) (networkmodel.ServerModel, error) {
	networks := make([]networkmodel.ServerNetwork, 0)
	if err := db.SelectContext(ctx, &networks, `
        SELECT uuid, server, network_name, ipv4_address FROM server_network WHERE server = $1
        `, model.UUID); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("faild to fetch networks: %w", err)
	}

	model.Networks = networks

	return model, nil
}

// fillServerModelHostPorts fetches the host ports configuration of a given server model from the database.
func fillServerModelHostPorts(ctx context.Context, db *sqlm.DB, model networkmodel.ServerModel) (networkmodel.ServerModel, error) {
	hostPorts := make([]networkmodel.HostPort, 0)
	if err := db.SelectContext(ctx, &hostPorts, `
        SELECT * FROM server_host_port WHERE server = $1
        `, model.UUID); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("faild to fetch server host ports: %w", err)
	}

	model.HostPorts = hostPorts

	return model, nil
}

// fillServerModelNetwork fetches the network configuration of a given server model from the database.
func fillServerModelOperator(ctx context.Context, db *sqlm.DB, model networkmodel.ServerModel) (networkmodel.ServerModel, error) {
	operator := networkmodel.ServerOperator{}
	if err := db.GetContext(ctx, &operator, `
        SELECT * FROM server_operator WHERE identifier = $1
        `, model.OperatorIdentifier); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("faild to fetch operator: %w", err)
	}

	model.OperatorRef = operator

	return model, nil
}
