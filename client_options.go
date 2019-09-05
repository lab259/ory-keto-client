package ketoclient

import (
	"github.com/gojek/heimdall/hystrix"
)

type Option func(*Client)

func New(opts ...Option) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	c._url = c.url.String()
	return c
}

// WithHystrixClient creates an option that will define the `hystrix.Client`
// when creating a new `Client`.
func WithHystrixClient(client *hystrix.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

// WithHost creates an option that defines a host name for the Keto server.
func WithHost(host string) Option {
	return func(c *Client) {
		c.url.Host = host
	}
}

// WithBaseURI creates an option that defines a base URI for the Keto server
// endpoints.
func WithBaseURI(baseURI string) Option {
	return func(c *Client) {
		c.url.Path = baseURI
	}
}
