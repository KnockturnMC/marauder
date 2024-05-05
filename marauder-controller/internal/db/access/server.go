package access

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"

	"gitea.knockturnmc.com/marauder/controller/sqlm"

	"github.com/google/uuid"
)

// FetchServer locates a server based on its uuid.
// sql.ErrNoRows is returned if no server exists with the passed uuid.
func FetchServer(ctx context.Context, db *sqlm.DB, uuid uuid.UUID) (networkmodel.ServerModel, error) {
	var result networkmodel.ServerModel
	if err := db.GetContext(ctx, &result, `
    SELECT * FROM server WHERE uuid = $1
    `, uuid); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to find server: %w", err)
	}

	return fillServerModelRefs(ctx, db, result)
}

// FetchServerByNameAndEnv looks up a single server based on its name and environment.
// sql.ErrNoRows is returned if no server exists with the passed name and environment.
func FetchServerByNameAndEnv(ctx context.Context, db *sqlm.DB, name string, environment string) (networkmodel.ServerModel, error) {
	var result networkmodel.ServerModel
	if err := db.GetContext(ctx, &result, `
    SELECT * FROM server WHERE name = $1 AND environment = $2
    `, name, environment); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to find server: %w", err)
	}

	return fillServerModelRefs(ctx, db, result)
}

// FetchServersByName queries the database for a collection of servers by their name.
func FetchServersByName(ctx context.Context, db *sqlm.DB, name string) ([]networkmodel.ServerModel, error) {
	result := make([]networkmodel.ServerModel, 0)
	if err := db.SelectContext(ctx, &result, `
    SELECT * FROM server WHERE name = $1
    `, name); err != nil {
		return result, fmt.Errorf("failed to find servers: %w", err)
	}

	return fillServerModelRefsSlice(ctx, db, result)
}

// FetchServersByEnvironment queries the database for a collection of servers by their environment.
func FetchServersByEnvironment(ctx context.Context, db *sqlm.DB, environment string) ([]networkmodel.ServerModel, error) {
	result := make([]networkmodel.ServerModel, 0)
	if err := db.SelectContext(ctx, &result, `
    SELECT * FROM server WHERE environment = $1
    `, environment); err != nil {
		return result, fmt.Errorf("failed to find servers: %w", err)
	}

	return fillServerModelRefsSlice(ctx, db, result)
}

// InsertServer creates a new server instance on the database.
func InsertServer(ctx context.Context, db *sqlm.DB, server networkmodel.ServerModel) (networkmodel.ServerModel, error) {
	transaction, err := db.Beginx()
	if err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() { _ = transaction.Rollback() }() // Rollback in case, this explodes. If Commit is called prior, this is a noop.

	if err := transaction.NamedGetContext(ctx, &server, `
            INSERT INTO server (environment, name, operator, memory, cpu, port, image)
            VALUES (:environment, :name, :operator, :memory, :cpu, :port, :image)
            RETURNING *; 
            `, server); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to insert server: %w", err)
	}

	for index, network := range server.Networks {
		network.ServerUUID = server.UUID // assign uuid generated from previous insertion.

		if err := transaction.NamedGetContext(ctx, &network, `
                INSERT INTO server_network (server, network_name, ipv4_address)
                VALUES (:server, :network_name, :ipv4_address)
                RETURNING *`,
			network); err != nil {
			return networkmodel.ServerModel{}, fmt.Errorf("failed to insert server network %s: %w", network.NetworkName, err)
		}

		server.Networks[index] = network
	}

	for index, hostPort := range server.HostPorts {
		hostPort.ServerUUID = server.UUID // assign uuid generated from previous insertion.

		if err := transaction.NamedGetContext(ctx, &hostPort, `
		            INSERT INTO server_host_port (server, host_ip, host_port, server_port)
					VALUES (:server, :host_ip, :host_port, :server_port)
					RETURNING *;
		        `, hostPort); err != nil {
			return networkmodel.ServerModel{}, fmt.Errorf("failed to insert host port %s:%d: %w", hostPort.HostIPAddr, hostPort.HostPort, err)
		}

		server.HostPorts[index] = hostPort
	}

	if server, err = fillServerModelOperator(ctx, db, server); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to fetch server operator of inserted server: %w", err)
	}

	if err := transaction.Commit(); err != nil {
		return networkmodel.ServerModel{}, fmt.Errorf("failed to commit insertion transaction: %w", err)
	}

	return server, nil
}
