package utils_test

import (
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Glob", Label("unittest"), func() {
	Describe("Finding the shortest glob match", func() {
		Context("for a valid path", func() {
			It("should find the shortest path", func() {
				cache := utils.NewShortestGlobPathCache()
				match, err := cache.FindShortestMatch("/var/local/spell*", "/var/local/spellcore/spells/hi.txt")

				Expect(err).To(Not(HaveOccurred()))
				Expect(match).To(BeEquivalentTo("/var/local/spellcore"))
			})
		})

		Context("for the root directory", func() {
			It("should find the shortest path, the root", func() {
				cache := utils.NewShortestGlobPathCache()
				match, err := cache.FindShortestMatch("/", "/var/local/spellcore/spells/hi.txt")

				Expect(err).To(Not(HaveOccurred()))
				Expect(match).To(BeEquivalentTo("/"))
			})
		})
	})
})
