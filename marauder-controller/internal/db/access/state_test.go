package access_test

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/internal/db/models"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("managing server state", func() {
	var (
		serverState models.ServerArtefactStateModel
		server      models.ServerModel
		artefact    models.ArtefactModel
	)

	BeforeEach(func() {
		databaseClient.MustExec("DELETE FROM server; DELETE FROM artefact;")

		var err error
		server, err = access.InsertServer(context.Background(), databaseClient, serverModel)
		Expect(err).To(Not(HaveOccurred()))
		artefact, err = access.InsertArtefact(context.Background(), databaseClient, fullArtefact)
		Expect(err).To(Not(HaveOccurred()))
		serverState = models.ServerArtefactStateModel{
			Server:         server.UUID,
			Artefact:       artefact.UUID,
			DefinitionDate: time.Now(),
			Type:           models.TARGET,
		}
	})

	Context("when inserting a new server state", func() {
		It("should properly insert the server state", func() {
			insertedModel, err := access.InsertServerState(context.Background(), databaseClient, serverState)
			Expect(err).To(Not(HaveOccurred()))

			var result models.ServerArtefactStateModel
			err = databaseClient.GetContext(context.Background(), &result, `
            SELECT uuid, server, artefact, definition_date, type FROM server_state WHERE uuid = $1`,
				insertedModel.UUID,
			)

			Expect(err).To(Not(HaveOccurred()))
			Expect(result).To(BeEquivalentTo(insertedModel))
		})

		It("should fail if a server state with the same name in the same environment is inserted twice", func() {
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
