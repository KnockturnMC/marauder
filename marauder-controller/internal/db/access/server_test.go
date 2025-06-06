package access_test

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var serverModel = networkmodel.ServerModel{
	Environment: "production",
	Name:        "hogwarts",
	OperatorRef: networkmodel.ServerOperator{
		Identifier: "falk0.servers.knockturnmc.com",
		Host:       "localhost",
		Port:       4200,
	},
	OperatorIdentifier: "falk0.servers.knockturnmc.com",
	Memory:             1024,
	CPU:                3,
	Port:               25565,
	Image:              "minecraft:paper",
	Networks: []networkmodel.ServerNetwork{
		{
			NetworkName: "services",
			IPV4Address: "172.18.0.4",
		},
	},
	HostPorts: make([]networkmodel.HostPort, 0),
}

var _ = Describe("managing servers", Label("functiontest"), func() {
	BeforeEach(func() {
		databaseClient.MustExec("DELETE FROM server_operator; DELETE FROM server; DELETE FROM server_network;")
		databaseClient.MustExec(fmt.Sprintf(
			"INSERT INTO server_operator VALUES ('%s', '%s', '%d')",
			serverModel.OperatorIdentifier,
			serverModel.OperatorRef.Host,
			serverModel.OperatorRef.Port,
		))
	})

	Context("when inserting a new server", func() {
		It("should properly insert the server", func() {
			insertedModel, err := access.InsertServer(context.Background(), databaseClient, serverModel)
			Expect(err).To(Not(HaveOccurred()))

			var result networkmodel.ServerModel
			err = databaseClient.GetContext(context.Background(), &result, `
            SELECT * FROM server WHERE uuid = $1`,
				insertedModel.UUID,
			)
			Expect(err).To(Not(HaveOccurred()))

			var networks []networkmodel.ServerNetwork
			err = databaseClient.SelectContext(context.Background(), &networks, `
            SELECT *
            FROM server_network
            WHERE server = $1`, insertedModel.UUID)
			Expect(err).To(Not(HaveOccurred()))

			result.Networks = networks

			var operator networkmodel.ServerOperator
			err = databaseClient.GetContext(context.Background(), &operator, `
			SELECT * FROM server_operator WHERE identifier = $1
			`, serverModel.OperatorIdentifier)
			Expect(err).To(Not(HaveOccurred()))

			result.OperatorRef = operator

			hostPorts := make([]networkmodel.HostPort, 0)
			err = databaseClient.SelectContext(context.Background(), &hostPorts, `
			SELECT * FROM server_host_port WHERE server = $1
			`, serverModel.UUID)
			Expect(err).To(Not(HaveOccurred()))

			result.HostPorts = hostPorts
			Expect(result).To(BeEquivalentTo(insertedModel))
		})

		It("should fail if a server with the same name in the same environment is inserted twice", func() {
			_, err := access.InsertServer(context.Background(), databaseClient, serverModel)
			Expect(err).To(Not(HaveOccurred()))

			_, err = access.InsertServer(context.Background(), databaseClient, serverModel)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when fetching a server by its uuid", func() {
		It("should find the server if if exists", func() {
			insertedModel, err := access.InsertServer(context.Background(), databaseClient, serverModel)
			Expect(err).To(Not(HaveOccurred()))

			server, err := access.FetchServer(context.Background(), databaseClient, insertedModel.UUID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(server).To(BeEquivalentTo(insertedModel))
		})

		It("should provide the correct error if no server exists", func() {
			_, err := access.FetchServer(context.Background(), databaseClient, uuid.New())
			Expect(err).To(MatchError(sql.ErrNoRows))
		})
	})

	Context("when fetching a server by name and environment", func() {
		It("should find the server if if exists", func() {
			insertedModel, err := access.InsertServer(context.Background(), databaseClient, serverModel)
			Expect(err).To(Not(HaveOccurred()))

			server, err := access.FetchServerByNameAndEnv(context.Background(), databaseClient, serverModel.Name, serverModel.Environment)
			Expect(err).To(Not(HaveOccurred()))
			Expect(server).To(BeEquivalentTo(insertedModel))
		})

		It("should provide the correct error if no server exists", func() {
			_, err := access.FetchServerByNameAndEnv(context.Background(), databaseClient, serverModel.Name, serverModel.Environment)
			Expect(err).To(MatchError(sql.ErrNoRows))
		})
	})

	Context("when fetching servers by their name", func() {
		It("should find all servers that exist", func() {
			insertionCount := 10
			for i := range insertionCount {
				server := serverModel
				server.Environment = strconv.Itoa(i)

				_, err := access.InsertServer(context.Background(), databaseClient, server)
				Expect(err).To(Not(HaveOccurred()))
			}

			servers, err := access.FetchServersByName(context.Background(), databaseClient, serverModel.Name)
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(servers)).To(BeEquivalentTo(insertionCount))

			for i, version := range servers {
				Expect(version.Environment).To(BeEquivalentTo(strconv.Itoa(i)))
			}
		})

		It("should return an empty slice if no servers exist", func() {
			servers, err := access.FetchServersByName(context.Background(), databaseClient, "crystals-home-server")
			Expect(err).To(Not(HaveOccurred()))
			Expect(servers).To(BeEmpty())
		})
	})

	Context("when fetching servers by their environment", func() {
		It("should find all servers that exist", func() {
			insertionCount := 10
			for i := range insertionCount {
				server := serverModel
				server.Name = strconv.Itoa(i)

				_, err := access.InsertServer(context.Background(), databaseClient, server)
				Expect(err).To(Not(HaveOccurred()))
			}

			servers, err := access.FetchServersByEnvironment(context.Background(), databaseClient, serverModel.Environment)
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(servers)).To(BeEquivalentTo(insertionCount))

			for i, version := range servers {
				Expect(version.Name).To(BeEquivalentTo(strconv.Itoa(i)))
			}
		})

		It("should return an empty slice if no servers exist", func() {
			servers, err := access.FetchServersByEnvironment(context.Background(), databaseClient, "crystals-home-environment")
			Expect(err).To(Not(HaveOccurred()))
			Expect(servers).To(BeEmpty())
		})
	})
})
