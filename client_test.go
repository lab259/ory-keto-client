package ketoclient_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/lab259/errors"
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

			response, err := client.AllowedOryAccessControlPolicy(ketoclient.Exact, &ketoclient.AllowedORYAccessControlPolicyRequest{
				Action:   "delete",
				Resource: "blog1:post:33",
				Subject:  "user:snake-eyes",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Allowed).To(BeTrue())
		})

		It("should deny an action that is not present", func() {
			client := ketoClient()

			response, err := client.AllowedOryAccessControlPolicy(ketoclient.Exact, &ketoclient.AllowedORYAccessControlPolicyRequest{
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

	Describe("ListOryAccessControlPolicy", func() {
		It("should list one policy", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.UpsertORYAccessPolicyRequest{
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

			listResponse, err := client.ListOryAccessControlPolicy(ketoclient.Exact, &ketoclient.ListORYAccessPolicyRequest{})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Policies).To(HaveLen(1))
			Expect(listResponse.Policies[0].ID).To(Equal("id1"))
			Expect(listResponse.Policies[0].Description).To(Equal("Delete action for Snake Eyes"))
			Expect(listResponse.Policies[0].Subjects).To(ConsistOf("user:snake-eyes"))
			Expect(listResponse.Policies[0].Actions).To(ConsistOf("delete"))
			Expect(listResponse.Policies[0].Resources).To(ConsistOf("blog1:post:33"))
			Expect(listResponse.Policies[0].Effect).To(Equal(ketoclient.Allow))
			Expect(listResponse.Policies[0].Conditions).To(HaveKeyWithValue("test", "value"))

		})

		It("should initialize a client with options", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.UpsertORYAccessPolicyRequest{
				ORYAccessControlPolicy: ketoclient.ORYAccessControlPolicy{
					ID:        "id1",
					Subjects:  []string{"user:snake-eyes"},
					Resources: []string{"blog1:post:33"},
					Actions:   []string{"delete"},
					Effect:    ketoclient.Allow,
					Conditions: map[string]interface{}{
						"test": "value",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			_, err = client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.UpsertORYAccessPolicyRequest{
				ORYAccessControlPolicy: ketoclient.ORYAccessControlPolicy{
					ID:        "id2",
					Subjects:  []string{"user:snake-eyes"},
					Resources: []string{"blog1:post:34"},
					Actions:   []string{"delete"},
					Effect:    ketoclient.Allow,
					Conditions: map[string]interface{}{
						"test": "value",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			_, err = client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.UpsertORYAccessPolicyRequest{
				ORYAccessControlPolicy: ketoclient.ORYAccessControlPolicy{
					ID:        "id3",
					Subjects:  []string{"user:snake-eyes"},
					Resources: []string{"blog1:post:35"},
					Actions:   []string{"delete"},
					Effect:    ketoclient.Allow,
					Conditions: map[string]interface{}{
						"test": "value",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			_, err = client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.UpsertORYAccessPolicyRequest{
				ORYAccessControlPolicy: ketoclient.ORYAccessControlPolicy{
					ID:        "id4",
					Subjects:  []string{"user:snake-eyes"},
					Resources: []string{"blog1:post:36"},
					Actions:   []string{"delete"},
					Effect:    ketoclient.Allow,
					Conditions: map[string]interface{}{
						"test": "value",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			_, err = client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.UpsertORYAccessPolicyRequest{
				ORYAccessControlPolicy: ketoclient.ORYAccessControlPolicy{
					ID:        "id5",
					Subjects:  []string{"user:snake-eyes"},
					Resources: []string{"blog1:post:37"},
					Actions:   []string{"delete"},
					Effect:    ketoclient.Allow,
					Conditions: map[string]interface{}{
						"test": "value",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			listResponse, err := client.ListOryAccessControlPolicy(ketoclient.Exact, &ketoclient.ListORYAccessPolicyRequest{
				Limit: 2,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Policies).To(HaveLen(2))
			Expect(listResponse.Policies[0].ID).To(Equal("id1"))
			Expect(listResponse.Policies[1].ID).To(Equal("id2"))

			listResponse, err = client.ListOryAccessControlPolicy(ketoclient.Exact, &ketoclient.ListORYAccessPolicyRequest{
				Limit:  2,
				Offset: 2,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Policies).To(HaveLen(2))
			Expect(listResponse.Policies[0].ID).To(Equal("id3"))
			Expect(listResponse.Policies[1].ID).To(Equal("id4"))

			listResponse, err = client.ListOryAccessControlPolicy(ketoclient.Exact, &ketoclient.ListORYAccessPolicyRequest{
				Limit:  2,
				Offset: 4,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Policies).To(HaveLen(1))
			Expect(listResponse.Policies[0].ID).To(Equal("id5"))
		})
	})

	Describe("GetOryAccessControlPolicy", func() {
		It("should get a policy", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.UpsertORYAccessPolicyRequest{
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

			response, err := client.GetOryAccessControlPolicy(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Policy.ID).To(Equal("id1"))
			Expect(response.Policy.Description).To(Equal("Delete action for Snake Eyes"))
			Expect(response.Policy.Subjects).To(ConsistOf("user:snake-eyes"))
			Expect(response.Policy.Actions).To(ConsistOf("delete"))
			Expect(response.Policy.Resources).To(ConsistOf("blog1:post:33"))
			Expect(response.Policy.Effect).To(Equal(ketoclient.Allow))
			Expect(response.Policy.Conditions).To(HaveKeyWithValue("test", "value"))
		})

		It("should fail getting a policy that does not exists", func() {
			client := ketoClient()

			response, err := client.GetOryAccessControlPolicy(ketoclient.Exact, "id1")
			Expect(err).To(HaveOccurred())
			Expect(response).To(BeNil())
			Expect(err).To(Equal(ketoclient.ErrNotFound))
		})
	})

	Describe("DeleteOryAccessControlPolicy", func() {
		It("should delete a policy", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlPolicy(ketoclient.Exact, &ketoclient.UpsertORYAccessPolicyRequest{
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

			err = client.DeleteOryAccessControlPolicy(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())

			listResponse, err := client.ListOryAccessControlPolicy(ketoclient.Exact, &ketoclient.ListORYAccessPolicyRequest{})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Policies).To(HaveLen(0))
		})

		It("should delete a non existing policy (YEASSSS, that is right)", func() {
			client := ketoClient()

			err := client.DeleteOryAccessControlPolicy(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("UpsertOryAccessControlRole", func() {
		It("should insert a role", func() {
			client := ketoClient()

			response, err := client.UpsertOryAccessControlRole(ketoclient.Exact, &ketoclient.UpsertORYAccessRoleRequest{
				Role: ketoclient.ORYAccessControlRole{
					ID:      "id1",
					Members: []string{"user:snake-eyes", "user:tank", "user:scarlet"},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())
			Expect(response.Role.ID).To(Equal("id1"))
			Expect(response.Role.Members).To(ConsistOf("user:snake-eyes", "user:tank", "user:scarlet"))

			getResponse, err := client.GetOryAccessControlRole(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())

			Expect(getResponse.Role.ID).To(Equal("id1"))
			Expect(getResponse.Role.Members).To(ConsistOf("user:snake-eyes", "user:tank", "user:scarlet"))
		})
	})

	Describe("GetOryAccessControlRole", func() {
		It("should get a role", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlRole(ketoclient.Exact, &ketoclient.UpsertORYAccessRoleRequest{
				Role: ketoclient.ORYAccessControlRole{
					ID:      "id1",
					Members: []string{"user:snake-eyes", "user:tank", "user:scarlet"},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			response, err := client.GetOryAccessControlRole(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())

			Expect(response.Role.ID).To(Equal("id1"))
			Expect(response.Role.Members).To(ConsistOf("user:snake-eyes", "user:tank", "user:scarlet"))
		})

		It("should fail getting a role that does not exists", func() {
			client := ketoClient()

			response, err := client.GetOryAccessControlRole(ketoclient.Exact, "id1")
			Expect(err).To(HaveOccurred())
			Expect(response).To(BeNil())
			Expect(err).To(Equal(ketoclient.ErrNotFound))
		})
	})

	Describe("ListOryAccessControlRole", func() {
		It("should list one role", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlRole(ketoclient.Exact, &ketoclient.UpsertORYAccessRoleRequest{
				Role: ketoclient.ORYAccessControlRole{
					ID:      "id1",
					Members: []string{"user:snake-eyes", "user:tank", "user:scarlet"},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			listResponse, err := client.ListOryAccessControlRole(ketoclient.Exact, &ketoclient.ListORYAccessRoleRequest{})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Roles).To(HaveLen(1))
			Expect(listResponse.Roles[0].ID).To(Equal("id1"))
			Expect(listResponse.Roles[0].Members).To(ConsistOf("user:snake-eyes", "user:tank", "user:scarlet"))

		})

		It("should initialize a client with options", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlRole(ketoclient.Exact, &ketoclient.UpsertORYAccessRoleRequest{
				Role: ketoclient.ORYAccessControlRole{
					ID:      "id1",
					Members: []string{"user1", "user2", "user3"},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			_, err = client.UpsertOryAccessControlRole(ketoclient.Exact, &ketoclient.UpsertORYAccessRoleRequest{
				Role: ketoclient.ORYAccessControlRole{
					ID:      "id2",
					Members: []string{"user3"},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			_, err = client.UpsertOryAccessControlRole(ketoclient.Exact, &ketoclient.UpsertORYAccessRoleRequest{
				Role: ketoclient.ORYAccessControlRole{
					ID:      "id3",
					Members: []string{"user4", "user5"},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			listResponse, err := client.ListOryAccessControlRole(ketoclient.Exact, &ketoclient.ListORYAccessRoleRequest{
				Limit: 2,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Roles).To(HaveLen(2))
			Expect(listResponse.Roles[0].ID).To(Equal("id1"))
			Expect(listResponse.Roles[1].ID).To(Equal("id2"))

			listResponse, err = client.ListOryAccessControlRole(ketoclient.Exact, &ketoclient.ListORYAccessRoleRequest{
				Limit:  2,
				Offset: 2,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Roles).To(HaveLen(1))
			Expect(listResponse.Roles[0].ID).To(Equal("id3"))
		})
	})

	Describe("DeleteOryAccessControlRole", func() {
		It("should delete a role", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlRole(ketoclient.Exact, &ketoclient.UpsertORYAccessRoleRequest{
				Role: ketoclient.ORYAccessControlRole{
					ID:      "id1",
					Members: []string{"user1", "user2", "user3"},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			err = client.DeleteOryAccessControlRole(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())

			listResponse, err := client.ListOryAccessControlRole(ketoclient.Exact, &ketoclient.ListORYAccessRoleRequest{})
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Roles).To(HaveLen(0))
		})

		It("should delete a non existing role (YEASSSS, that is right)", func() {
			client := ketoClient()

			err := client.DeleteOryAccessControlRole(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("AddMembersOryAccessControlRole", func() {
		It("should add some members to a role", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlRole(ketoclient.Exact, &ketoclient.UpsertORYAccessRoleRequest{
				Role: ketoclient.ORYAccessControlRole{
					ID:      "id1",
					Members: []string{"snake-eyes"},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			response, err := client.AddMembersOryAccessControlRole(ketoclient.Exact, "id1", &ketoclient.AddMembersORYAccessRoleRequest{
				Members: []string{"scarlet", "tank"},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Role.ID).To(Equal("id1"))
			Expect(response.Role.Members).To(ConsistOf("snake-eyes", "scarlet", "tank"))

			getResponse, err := client.GetOryAccessControlRole(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())
			Expect(getResponse.Role.ID).To(Equal("id1"))
			Expect(getResponse.Role.Members).To(ConsistOf("snake-eyes", "scarlet", "tank"))
		})

		It("should add members to a non existing role (YEASSSS, that is right)", func() {
			client := ketoClient()

			response, err := client.AddMembersOryAccessControlRole(ketoclient.Exact, "id1", &ketoclient.AddMembersORYAccessRoleRequest{
				Members: []string{"scarlet", "tank"},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Role.ID).To(Equal("id1"))
			Expect(response.Role.Members).To(ConsistOf("scarlet", "tank"))

			getResponse, err := client.GetOryAccessControlRole(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())
			Expect(getResponse.Role.ID).To(Equal("id1"))
			Expect(getResponse.Role.Members).To(ConsistOf("scarlet", "tank"))
		})
	})

	Describe("RemoveMemberOryAccessControlRole", func() {
		It("should remove a member from a rule", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlRole(ketoclient.Exact, &ketoclient.UpsertORYAccessRoleRequest{
				Role: ketoclient.ORYAccessControlRole{
					ID:      "id1",
					Members: []string{"snake-eyes", "scarlet"},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			err = client.RemoveMemberOryAccessControlRole(ketoclient.Exact, "id1", "scarlet")
			Expect(err).ToNot(HaveOccurred())

			getResponse, err := client.GetOryAccessControlRole(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())
			Expect(getResponse.Role.ID).To(Equal("id1"))
			Expect(getResponse.Role.Members).To(ConsistOf("snake-eyes"))
		})

		It("should remove members that are not in the role", func() {
			client := ketoClient()

			_, err := client.UpsertOryAccessControlRole(ketoclient.Exact, &ketoclient.UpsertORYAccessRoleRequest{
				Role: ketoclient.ORYAccessControlRole{
					ID:      "id1",
					Members: []string{"snake-eyes"},
				},
			})
			Expect(err).ToNot(HaveOccurred())

			err = client.RemoveMemberOryAccessControlRole(ketoclient.Exact, "id1", "scarlet")
			Expect(err).ToNot(HaveOccurred())

			getResponse, err := client.GetOryAccessControlRole(ketoclient.Exact, "id1")
			Expect(err).ToNot(HaveOccurred())
			Expect(getResponse.Role.ID).To(Equal("id1"))
			Expect(getResponse.Role.Members).To(ConsistOf("snake-eyes"))
		})
	})

	Describe("HealthAlive", func() {
		It("should the server be alive", func() {
			client := ketoClient()

			response, err := client.HealthAlive()
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Status).To(Equal("ok"))
		})
	})

	Describe("HealthReadness", func() {
		It("should the server be ready", func() {
			client := ketoClient()

			response, err := client.HealthReadness()
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Status).To(Equal("ok"))
		})
	})

	Describe("Version", func() {
		It("should return the server version", func() {
			client := ketoClient()

			response, err := client.Version()
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Version).To(Equal("v0.3.3-sandbox+oryOS.12"))
		})
	})

	Describe("CheckVersion", func() {
		It("should check if the server is compatible with this library version", func() {
			client := ketoClient()

			Expect(client.CheckVersion()).To(Succeed())
		})

		It("should fail checking the server compatibility with an old server", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, req *http.Request) {
				response.Write([]byte(`{"version":"v0.2.1"}`))
			}))
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			Expect(err).ToNot(HaveOccurred())
			client := ketoclient.New(ketoclient.WithURL(u))

			err = client.CheckVersion()
			Expect(errors.Is(err, ketoclient.ErrServerIncompatible)).To(BeTrue())
			Expect(err.Error()).To(HavePrefix("got v0.2.1"))
		})
	})
})
