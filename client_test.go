package ketoclient_test

import (
	"net/url"

	ketoclient "github.com/lab259/ory-keto-client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func ketoURL() *url.URL {
	u, err := url.Parse("http://localhost:4466")
	if err != nil {
		panic(err)
	}
	return u
}

func ketoClient() *ketoclient.Client {
	return ketoclient.New(
		ketoclient.WithURL(ketoURL()),
	)
}

var _ = Describe("Client", func() {
	Describe("Allowed", func() {
		It("should allow an action", func() {
			client := ketoClient()
			_, err := client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.AcpUpsertORYAccessPolicyRequest{
				ORYAccessControlPolicy: ketoclient.ORYAccessControlPolicy{
					ID:          "id1",
					Description: "Delete action for Snake Eyes",
					Subjects:    []string{"user:snake-eyes"},
					Resources:   []string{"blog1:post:33"},
					Actions:     []string{"delete"},
					Effect:      ketoclient.Allow,
				},
			})
			Expect(err).ToNot(HaveOccurred())

			response, err := client.Allowed(ketoclient.Exact, &ketoclient.AcpAllowedRequest{
				Action:   "delete",
				Resource: "blog1:post:33",
				Subject:  "user:snake-eyes",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Allowed).To(BeTrue())
		})

		It("should deny an action that is not present", func() {
			client := ketoClient()

			response, err := client.Allowed(ketoclient.Exact, &ketoclient.AcpAllowedRequest{
				Action:   "delete",
				Resource: "blog1:post:34",
				Subject:  "user:snake-eyes",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Allowed).To(BeFalse())
		})
	})

	Describe("UpsertOryAccessControlPolicy", func() {
		It("should initialize a client with options", func() {
			client := ketoClient()

			response, err := client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.AcpUpsertORYAccessPolicyRequest{
				ORYAccessControlPolicy: ketoclient.ORYAccessControlPolicy{
					ID:          "id1",
					Description: "Delete action for Snake Eyes",
					Subjects:    []string{"user:snake-eyes"},
					Resources:   []string{"blog1:post:33"},
					Actions:     []string{"delete"},
					Effect:      ketoclient.Allow,
					Conditions: map[string]interface{}{
						"test": "value",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())
			Expect(response.ID).To(Equal("id1"))
			Expect(response.Description).To(Equal("Delete action for Snake Eyes"))
			Expect(response.Subjects).To(ConsistOf("user:snake-eyes"))
			Expect(response.Actions).To(ConsistOf("delete"))
			Expect(response.Resources).To(ConsistOf("blog1:post:33"))
			Expect(response.Effect).To(Equal(ketoclient.Allow))
			Expect(response.Conditions).To(HaveKeyWithValue("test", "value"))

			// TODO To get the list
		})
	})
})
