package ketoclient

import (
	"net/url"

	"github.com/gojek/heimdall/hystrix"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	It("should initialize a client without hystrix client", func() {
		client := New()
		Expect(client).ToNot(BeNil())
	})

	It("should initialize a client with options", func() {
		hc := hystrix.NewClient()
		url, err := url.Parse("http://host1/baseURI")
		Expect(err).ToNot(HaveOccurred())
		client := New(
			WithURL(url),
			WithHystrixClient(hc),
		)
		Expect(client).ToNot(BeNil())
		Expect(client._url).To(Equal("http://host1/baseURI"))
		Expect(client.client).To(Equal(hc))
	})
})
