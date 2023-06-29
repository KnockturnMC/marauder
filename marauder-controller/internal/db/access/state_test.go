package access_test

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/google/uuid"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/internal/db/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("managing server state", Label("functiontest"), func() {
	var (
		serverState models.ServerArtefactStateModel
		server      models.ServerModel
		artefact    models.ArtefactModel
	)

	BeforeEach(func() {
		databaseClient.MustExec("DELETE FROM server; DELETE FROM server_network; DELETE FROM artefact;")

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

		It("should fail if the server has the same is state and is inserting an is state", func() {
			serverState.Type = models.IS

			_, err := access.InsertServerState(context.Background(), databaseClient, serverState)
			Expect(err).To(Not(HaveOccurred()))

			_, err = access.InsertServerState(context.Background(), databaseClient, serverState)
			Expect(err).To(HaveOccurred())
		})

		It("should fail if the server has the same target state and is inserting an target state", func() {
			serverState.Type = models.TARGET

			_, err := access.InsertServerState(context.Background(), databaseClient, serverState)
			Expect(err).To(Not(HaveOccurred()))

			_, err = access.InsertServerState(context.Background(), databaseClient, serverState)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when querying a servers current is/target state", func() {
		type AccessMethod func(ctx context.Context, db *sqlm.DB, serverUUID uuid.UUID) (models.ServerArtefactStateModel, error)
		for serverStateType, fetchMethod := range map[models.ServerStateType]AccessMethod{
			models.IS:     access.FetchServerIsState,
			models.TARGET: access.FetchServerTargetState,
		} {
			Context(fmt.Sprintf("for the %s state", serverStateType), func() {
				It("should properly fetch the state if it exists", func() {
					serverState.Type = serverStateType
					state, err := access.InsertServerState(context.Background(), databaseClient, serverState)
					Expect(err).To(Not(HaveOccurred()))

					model, err := fetchMethod(context.Background(), databaseClient, server.UUID)
					Expect(err).To(Not(HaveOccurred()))

					Expect(model).To(BeEquivalentTo(state))
				})

				It("should properly error if the state does not exist", func() {
					_, err := fetchMethod(context.Background(), databaseClient, server.UUID)
					Expect(err).To(MatchError(sql.ErrNoRows))
				})
			})
		}
	})
})
