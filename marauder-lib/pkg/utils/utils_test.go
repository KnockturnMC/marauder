package utils_test

import (
	"bytes"

	. "github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", Label("unittest"), func() {
	Describe("Converting an int to and from bytes", func() {
		It("works fine for positive numbers", func() {
			slice := IntToByteSlice(15)
			toInt, err := ByteSliceToInt(bytes.NewBuffer(slice))

			Expect(err).To(Not(HaveOccurred()))
			Expect(toInt).To(BeEquivalentTo(15))
		})

		It("works fine for 0", func() {
			slice := IntToByteSlice(0)
			toInt, err := ByteSliceToInt(bytes.NewBuffer(slice))

			Expect(err).To(Not(HaveOccurred()))
			Expect(toInt).To(BeEquivalentTo(0))
		})

		It("works fine for negative numbers", func() {
			slice := IntToByteSlice(-103)
			toInt, err := ByteSliceToInt(bytes.NewBuffer(slice))

			Expect(err).To(Not(HaveOccurred()))
			Expect(toInt).To(BeEquivalentTo(-103))
		})
	})
})
