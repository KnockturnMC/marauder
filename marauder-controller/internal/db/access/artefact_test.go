package access_test

import (
	"context"
	"time"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"gitea.knockturnmc.com/marauder/controller/internal/db/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("managing artefacts on the db", Label("functiontest"), func() {
	Context("inserting a proper artefact into the database", func() {
		It("should properly insert it", func() {
			artefact, err := access.InsertArtefact(context.Background(), databaseClient, models.ArtefactModelWithBinary{
				ArtefactModel: models.ArtefactModel{
					Identifier: "spellcore",
					Version:    "1.0.0+hello",
					UploadDate: time.Now(),
				},
				TarballBlob: []byte("example data"),
			})

			Expect(err).To(Not(HaveOccurred()))
			Expect(artefact.UUID).To(Not(BeEmpty()))

			var testModel models.ArtefactModel
			err = databaseClient.Get(&testModel, `SELECT * FROM artefact WHERE uuid = $1`, artefact.UUID)

			Expect(err).To(Not(HaveOccurred()))
			Expect(testModel).To(BeEquivalentTo(artefact))
		})
	})
})
