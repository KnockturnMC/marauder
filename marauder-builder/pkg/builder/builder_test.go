package builder_test

import (
	"testing/fstest"

	"gitea.knockturnmc.com/marauder/builder/pkg/builder"

	"gitea.knockturnmc.com/marauder/lib/pkg/artefact"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils"
	"gitea.knockturnmc.com/marauder/lib/pkg/utils/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Building the artefact", func() {
	Describe("copying the files described into a tarball", func() {
		var rootFS fstest.MapFS

		BeforeEach(func() {
			rootFS = make(fstest.MapFS)
		})

		Context("for existing files", func() {
			It("should properly write to the tar writer", func() {
				rootFS["spell-plugin/build/libs/spellcore-1.14.jar"] = &fstest.MapFile{Data: []byte("plugin")}
				rootFS["spell-api/build/libs/spellbook-1.14.jar"] = &fstest.MapFile{Data: []byte("api")}

				writer := mocks.NewFriendlyTarballWriter(GinkgoT())
				writer.EXPECT().AddFile(mock.Anything, "spell-plugin/build/libs/spellcore-1.14.jar", "files/spellcore.jar").Return(nil)
				writer.EXPECT().AddFile(mock.Anything, "spell-api/build/libs/spellbook-1.14.jar", "files/spellapi.jar").Return(nil)

				err := builder.IncludeArtefactFiles(&rootFS, artefact.Manifest{
					Identifier: "spellcore",
					Version:    "1.14",
					Files: []artefact.FileReference{
						{Target: "spellcore.jar", CISourceGlob: "spell-plugin/build/libs/spellcore-*.jar"},
						{Target: "spellapi.jar", CISourceGlob: "spell-api/build/libs/spellbook-*.jar"},
					},
				}, utils.NewShortestGlobPathCache(), writer)

				Expect(err).To(Not(HaveOccurred()))
			})
		})
	})
})
