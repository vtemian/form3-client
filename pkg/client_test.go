package pkg

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vtemian/form3/pkg/api"
)

var _ = Describe("Form3Client", func() {
	form3Client, _ := NewClient()

	Describe("Fetch account", func() {
		It("should return an account", func() {
			account := &api.Account{}

			err := form3Client.Fetch(context.TODO(), account)
			Expect(err).To(BeNil())
		})
	})
})
