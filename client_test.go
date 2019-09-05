package ketoclient_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	ketoclient "github.com/lab259/ory-keto-client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ory/dockertest"
)

var ketoServicePort string

func ketoURL() *url.URL {
	u, err := url.Parse(fmt.Sprintf("http://localhost:%s", ketoServicePort))
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
	var (
		pool     *dockertest.Pool
		resource *dockertest.Resource
	)

	BeforeEach(func() {
		var err error

		logger := log.New(GinkgoWriter, "[keto initialization] ", 0)

		logger.Println("Starting pool")
		pool, err = dockertest.NewPool("")
		pool.MaxWait = time.Second * 10
		Expect(err).ToNot(HaveOccurred())

		logger.Println("Running keto")
		resource, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository:   "oryd/keto",
			Tag:          "v0.3.3-sandbox",
			Env:          []string{"DSN=memory"},
			Cmd:          []string{"serve"},
			ExposedPorts: []string{"4466"},
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(pool.Retry(func() error {
			ketoServicePort = resource.GetPort("4466/tcp")
			serviceURL := fmt.Sprintf("http://localhost:%s/version", ketoServicePort)

			response, err := http.DefaultClient.Get(serviceURL)
			if err != nil {
				return err
			}
			defer response.Body.Close()
			data, err := ioutil.ReadAll(response.Body)
			logger.Println("Keto service: ", response.Status, ":", string(data))
			return nil
		})).To(Succeed())
	})

	AfterEach(func() {
		Expect(pool.Purge(resource)).To(Succeed())
	})

	Describe("Allowed", func() {
		It("should allow an action", func() {
			client := ketoClient()
			_, err := client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.UpsertORYAccessPolicyRequest{
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

			response, err := client.Allowed(ketoclient.Exact, &ketoclient.AllowedORYAccessControlPolicyRequest{
				Action:   "delete",
				Resource: "blog1:post:33",
				Subject:  "user:snake-eyes",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Allowed).To(BeTrue())
		})

		It("should deny an action that is not present", func() {
			client := ketoClient()

			response, err := client.Allowed(ketoclient.Exact, &ketoclient.AllowedORYAccessControlPolicyRequest{
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

			response, err := client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.UpsertORYAccessPolicyRequest{
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

			listResponse, err := client.ListOryAccessControlPolicy(ketoclient.Exact, &ketoclient.ListORYAccessPolicyRequest{})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Policies).To(HaveLen(1))
			Expect(listResponse.Policies[0].ID).To(Equal("id1"))
		})
	})
})
