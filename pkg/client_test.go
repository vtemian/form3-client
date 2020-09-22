package pkg

import (
	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
)

var _ = Describe("Form3Client", func() {
	It("should throw error", func() {
		Expect("test").To(Equal("not-test"))
	})
})
