package access_test

import (
	"context"
	"database/sql"
	"encoding/hex"
	"strconv"
	"time"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func MustHexDecode(val string) []byte {
	decodeString, err := hex.DecodeString(val)
	if err != nil {
		panic(err)
	}

	return decodeString
}

var fullArtefact = networkmodel.ArtefactModelWithBinary{
	ArtefactModel: networkmodel.ArtefactModel{
		Identifier:      "spellcore",
		Version:         "1.0.0+hello",
		UploadDate:      time.Now(),
		RequiresRestart: true,
	},
	TarballBlob: []byte("example data"),
	Hash:        MustHexDecode("ebbc0ce59cea35533cfb2d63443fb3db650e9d263ba3f91aee70110a108a6ff9"),
}

var _ = Describe("managing artefacts on the db", Label("functiontest"), func() {
	BeforeEach(func() {
		databaseClient.MustExec("DELETE FROM artefact")
	})

	Context("inserting a proper artefact into the database", func() {
		It("should properly insert it", func() {
			artefact, err := access.InsertArtefact(context.Background(), databaseClient, fullArtefact)

			Expect(err).To(Not(HaveOccurred()))
			Expect(artefact.UUID).To(Not(BeEmpty()))

			var testModel networkmodel.ArtefactModel
			err = databaseClient.Get(&testModel, `SELECT * FROM artefact WHERE uuid = $1`, artefact.UUID)

			Expect(err).To(Not(HaveOccurred()))
			Expect(testModel).To(BeEquivalentTo(artefact))

			tarballRow, err := databaseClient.Queryx(`SELECT tarball FROM artefact_file WHERE artefact = $1`, testModel.UUID)
			Expect(err).To(Not(HaveOccurred()))
			defer func() { _ = tarballRow.Close() }()

			Expect(tarballRow.Next()).To(BeTrue())
			res := map[string]interface{}{}
			Expect(tarballRow.MapScan(res)).To(Not(HaveOccurred()))
			Expect(res["tarball"]).To(BeEquivalentTo([]byte("example data")))
		})

		It("should properly fail if the artefact identifier and version combination already exits", func() {
			_, err := access.InsertArtefact(context.Background(), databaseClient, fullArtefact)
			Expect(err).To(Not(HaveOccurred()))

			_, errOnDuplicateInsertion := access.InsertArtefact(context.Background(), databaseClient, fullArtefact)
			Expect(errOnDuplicateInsertion).To(HaveOccurred())
		})
	})

	Context("when fetching a artefact based on its identifier and version", func() {
		It("should find the a single artefact if only one exists", func() {
			insertedArtefact, err := access.InsertArtefact(context.Background(), databaseClient, fullArtefact)
			Expect(err).To(Not(HaveOccurred()))

			artefact, err := access.FetchArtefact(context.Background(), databaseClient, fullArtefact.Identifier, fullArtefact.Version)
			Expect(err).To(Not(HaveOccurred()))
			Expect(artefact).To(BeEquivalentTo(insertedArtefact))
		})

		It("should return the proper error when no data is present", func() {
			_, err := access.FetchArtefact(context.Background(), databaseClient, fullArtefact.Identifier, fullArtefact.Version)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(sql.ErrNoRows))
		})
	})

	Context("when fetching an artefact based purely on its identifier", func() {
		It("should find all registered artefacts with the given identifier", func() {
			insertionCount := 10
			for i := 0; i < insertionCount; i++ {
				artefact := fullArtefact
				artefact.Version = strconv.Itoa(i)

				_, err := access.InsertArtefact(context.Background(), databaseClient, artefact)
				Expect(err).To(Not(HaveOccurred()))
			}

			versions, err := access.FetchArtefactVersions(context.Background(), databaseClient, fullArtefact.Identifier)
			Expect(err).To(Not(HaveOccurred()))
			Expect(len(versions)).To(BeEquivalentTo(insertionCount))

			for i, version := range versions {
				Expect(version.Version).To(BeEquivalentTo(strconv.Itoa(i)))
			}
		})

		It("should return an empty slice if no artefacts exist", func() {
			versions, err := access.FetchArtefactVersions(context.Background(), databaseClient, fullArtefact.Identifier)
			Expect(err).To(Not(HaveOccurred()))
			Expect(versions).To(BeEmpty())
		})
	})

	Context("when fetching the tarball for a specific database", func() {
		It("should find the tarball for an existing artefact", func() {
			artefact, err := access.InsertArtefact(context.Background(), databaseClient, fullArtefact)
			Expect(err).To(Not(HaveOccurred()))

			tarball, err := access.FetchArtefactTarball(context.Background(), databaseClient, artefact.UUID)
			Expect(err).To(Not(HaveOccurred()))
			Expect(tarball.TarballBlob).To(BeEquivalentTo("example data"))
		})

		It("should return the proper error if no artefact tarball is found", func() {
			_, err := access.FetchArtefactTarball(context.Background(), databaseClient, uuid.New())
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(sql.ErrNoRows))
		})
	})
})
