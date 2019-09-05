package ketoclient

import (
	"net/url"

	"github.com/gojek/heimdall/hystrix"
)

type Option func(*Client)

func New(opts ...Option) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	if c.url.Scheme == "" {
		c.url.Scheme = "http"
	}
	c._url = c.url.String()
	if c.client == nil {
		c.client = hystrix.NewClient()
	}
	return c
}

// WithHystrixClient creates an option that will define the `hystrix.Client`
// when creating a new `Client`.
func WithHystrixClient(client *hystrix.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

// WithURL creates an option that defines a host name for the Keto server.
func WithURL(u *url.URL) Option {
	return func(c *Client) {
		c.url = *u
	}
}
