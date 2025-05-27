package builder_test

import (
	"io/fs"
	"testing/fstest"

	"github.com/stretchr/testify/mock"

	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils/mocks"

	"github.com/knockturnmc/marauder/marauder-client/pkg/builder"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/filemodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/mo"
)

var _ = Describe("Building the artefact", Label("unittest"), func() {
	Describe("copying the files described into a tarball", func() {
		var rootFS fstest.MapFS
		okayTarballResponseSingleFile := func(_ fs.FS, pathOnDisk string, pathInTarball string) ([]utils.WriteResult, error) {
			return []utils.WriteResult{{
				PathInRootFS:  pathOnDisk,
				PathInTarball: mo.Some(pathInTarball),
			}}, nil
		}

		BeforeEach(func() {
			rootFS = make(fstest.MapFS)
		})

		Context("for existing files", func() {
			It("should properly write to the tar writer", func() {
				rootFS["spell-plugin/build/libs/spellcore-1.14.jar"] = &fstest.MapFile{Data: []byte("plugin")}
				rootFS["spell-api/build/libs/spellbook-1.14.jar"] = &fstest.MapFile{Data: []byte("api")}

				writer := mocks.NewMockFriendlyTarballWriter(GinkgoT())
				writer.On("WithFilter", mock.Anything).Return(writer)
				writer.On("Add", mock.Anything, "spell-plugin/build/libs/spellcore-1.14.jar", "files/spellcore.jar").
					Return(okayTarballResponseSingleFile)
				writer.On("Add", mock.Anything, "spell-api/build/libs/spellbook-1.14.jar", "files/spellapi.jar").
					Return(okayTarballResponseSingleFile)

				_, err := builder.IncludeArtefactFiles(&rootFS, filemodel.Manifest{
					Identifier: "spellcore",
					Version:    "1.14",
					Files: filemodel.FileReferenceCollection{
						{Target: "spellcore.jar", CISourceGlob: "spell-plugin/build/libs/spellcore-*.jar"},
						{Target: "spellapi.jar", CISourceGlob: "spell-api/build/libs/spellbook-*.jar"},
					},
				}, utils.NewShortestGlobPathCache(), writer)

				Expect(err).To(Not(HaveOccurred()))
			})
		})
	})
})
