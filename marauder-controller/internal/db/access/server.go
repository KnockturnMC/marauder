package access

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/networkmodel"

	"gitea.knockturnmc.com/marauder/controller/sqlm"

	"github.com/google/uuid"
)

// FetchServer locates a server based on its uuid.
// sql.ErrNoRows is returned if no server exists with the passed uuid.
func FetchServer(ctx context.Context, db *sqlm.DB, uuid uuid.UUID) (networkmodel.ServerModel, error) {
	var result networkmodel.ServerModel
	if err := db.GetContext(ctx, &result, `
    SELECT uuid, environment, name, host, memory, cpu, image FROM server WHERE uuid = $1
    `, uuid); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to find server: %w", err)
	}

	return fillServerModelNetwork(ctx, db, result)
}

// FetchServerByNameAndEnv looks up a single server based on its name and environment.
// sql.ErrNoRows is returned if no server exists with the passed name and environment.
func FetchServerByNameAndEnv(ctx context.Context, db *sqlm.DB, name string, environment string) (networkmodel.ServerModel, error) {
	var result networkmodel.ServerModel
	if err := db.GetContext(ctx, &result, `
    SELECT uuid, environment, name, host, memory, cpu, image FROM server WHERE name = $1 AND environment = $2
    `, name, environment); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to find server: %w", err)
	}

	return fillServerModelNetwork(ctx, db, result)
}

// FetchServersByName queries the database for a collection of servers by their name.
func FetchServersByName(ctx context.Context, db *sqlm.DB, name string) ([]networkmodel.ServerModel, error) {
	var result []networkmodel.ServerModel
	if err := db.SelectContext(ctx, &result, `
    SELECT uuid, environment, name, host, memory, cpu, image FROM server WHERE name = $1
    `, name); err != nil {
		return result, fmt.Errorf("failed to find servers: %w", err)
	}

	return fillServerModelNetworkSlice(ctx, db, result)
}

// FetchServersByEnvironment queries the database for a collection of servers by their environment.
func FetchServersByEnvironment(ctx context.Context, db *sqlm.DB, environment string) ([]networkmodel.ServerModel, error) {
	var result []networkmodel.ServerModel
	if err := db.SelectContext(ctx, &result, `
    SELECT uuid, environment, name, host, memory, cpu, image FROM server WHERE environment = $1
    `, environment); err != nil {
		return result, fmt.Errorf("failed to find servers: %w", err)
	}

	return fillServerModelNetworkSlice(ctx, db, result)
}

// fillServerModelNetwork fetches the network configuration of a given server model from the database.
func fillServerModelNetwork(ctx context.Context, db *sqlm.DB, model networkmodel.ServerModel) (networkmodel.ServerModel, error) {
	var networks []networkmodel.ServerNetwork
	if err := db.SelectContext(ctx, &networks, `
        SELECT uuid, server, network_name, ipv4_address FROM server_network WHERE server = $1
        `, model.UUID); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("faild to fetch networks: %w", err)
	}

	model.Networks = networks

	return model, nil
}

// fillServerModelNetworkSlice executes fillServerModelNetwork for each server in the slice.
func fillServerModelNetworkSlice(ctx context.Context, db *sqlm.DB, servers []networkmodel.ServerModel) ([]networkmodel.ServerModel, error) {
	for i, el := range servers {
		fullyFetched, err := fillServerModelNetwork(ctx, db, el)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch fullyFetched for %s: %w", el.UUID.String(), err)
		}
		servers[i] = fullyFetched
	}

	return servers, nil
}

// InsertServer creates a new server instance on the database.
func InsertServer(ctx context.Context, db *sqlm.DB, server networkmodel.ServerModel) (networkmodel.ServerModel, error) {
	transaction, err := db.Beginx()
	if err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = transaction.Rollback() }() // Rollback in case, this explodes. If Commit is called prior, this is a noop.

	if err := transaction.NamedGetContext(ctx, &server, `
            INSERT INTO server (environment, name, host, memory, cpu, image)
            VALUES (:environment, :name, :host, :memory, :cpu, :image)
            RETURNING *; 
            `, server); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to insert server: %w", err)
	}

	for index, network := range server.Networks {
		result := network               // avoid pointer creating to for loop parameter
		result.ServerUUID = server.UUID // assign uuid generated from previous insertion.

		if err := transaction.NamedGetContext(ctx, &result, `
                INSERT INTO server_network (server, network_name, ipv4_address) 
                VALUES (:server, :network_name, :ipv4_address)
                RETURNING *`,
			result); err != nil {
			return networkmodel.ServerModel{}, fmt.Errorf("failed to insert server network %s: %w", network.NetworkName, err)
		}

		server.Networks[index] = result
	}

	if err := transaction.Commit(); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to commit insertion transaction: %w", err)
	}

	return server, nil
}
