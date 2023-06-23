package access_test

import (
	"context"
	"database/sql"
	"strconv"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/internal/db/models"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var serverModel = models.ServerModel{
	Environment: "production",
	Name:        "hogwarts",
	Host:        "falk0.servers.knockturnmc.com",
	Memory:      1024,
	Image:       "minecraft:paper",
}

var _ = Describe("managing servers", func() {
	BeforeEach(func() {
		databaseClient.MustExec("DELETE FROM server;")
	})

	Context("when inserting a new server", func() {
		It("should properly insert the server", func() {
			insertedModel, err := access.InsertServer(context.Background(), databaseClient, serverModel)
			Expect(err).To(Not(HaveOccurred()))

			var result models.ServerModel
			err = databaseClient.GetContext(context.Background(), &result, `
            SELECT uuid, environment, name, host, memory, image FROM server WHERE uuid = $1`,
				insertedModel.UUID,
			)

			Expect(err).To(Not(HaveOccurred()))
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
			for i := 0; i < insertionCount; i++ {
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
			for i := 0; i < insertionCount; i++ {
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
