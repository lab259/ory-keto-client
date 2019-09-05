package ketoclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gojek/heimdall/hystrix"
)

type UnexpectedResponse struct {
	Response *http.Response
}

func (err *UnexpectedResponse) Error() string {
	return fmt.Sprintf("unexpected status %s", http.StatusText(err.Response.StatusCode))
}

type Client struct {
	url    url.URL
	_url   string
	client *hystrix.Client
}

type Flavor string

const (
	// Exact does an case sensitive equality when comparing the rules.
	//
	// See Also https://www.ory.sh/docs/keto/engines/acp-ory#pattern-matching-strategies
	Exact Flavor = "exact"

	// Glob uses more advanced matching. It supports wildcards, single symbol
	// wildcards, super wildcards, character lists, etc.
	//
	// See Also https://www.ory.sh/docs/keto/engines/acp-ory#pattern-matching-strategies
	Glob Flavor = "glob"

	// Regex uses regexp to match the rules.
	//
	// See Also https://www.ory.sh/docs/keto/engines/acp-ory#pattern-matching-strategies
	Regex Flavor = "regex"
)

// Allowed check if a request is allowed.
//
// See Also https://www.ory.sh/docs/keto/sdk/api#check-if-a-request-is-allowed
func (client *Client) Allowed(flavor Flavor, request *AcpAllowedRequest) (*AcpAllowedResponse, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err := enc.Encode(request)
	if err != nil {
		return nil, err
	}

	response, err := client.client.Post(client._url+"/"+string(flavor)+"/allowed", buf, nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		return &AcpAllowedResponse{Allowed: true}, nil
	case http.StatusForbidden:
		return &AcpAllowedResponse{Allowed: false}, nil
	case http.StatusInternalServerError:
		r := &ResponseError{}
		dec := json.NewDecoder(response.Body)
		err := dec.Decode(r)
		if err != nil {
			return nil, err
		}
		return nil, r
	default:
		return nil, &UnexpectedResponse{Response: response}
	}
}
