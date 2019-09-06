package ketoclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/lab259/errors/v2"

	"github.com/blang/semver"
	"github.com/gojek/heimdall/hystrix"
)

const clientVersionCompatibility = ">=0.3.0"

var (
	clientVersion = semver.MustParseRange(clientVersionCompatibility)

	ErrNotFound           = errors.New("policy not found")
	ErrServerIncompatible = errors.New("server incompatible. required: " + clientVersionCompatibility)
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

// AllowedOryAccessControlPolicy check if a request is allowed.
//
// See Also https://www.ory.sh/docs/keto/sdk/api#check-if-a-request-is-allowed
func (client *Client) AllowedOryAccessControlPolicy(flavor Flavor, request *AllowedORYAccessControlPolicyRequest) (*AllowedORYAccessControlPolicyResponse, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err := enc.Encode(request)
	if err != nil {
		return nil, err
	}

	response, err := client.client.Post(client._url+"/engines/acp/ory/"+string(flavor)+"/allowed", buf, nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		return &AllowedORYAccessControlPolicyResponse{Allowed: true}, nil
	case http.StatusForbidden:
		return &AllowedORYAccessControlPolicyResponse{Allowed: false}, nil
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

// UpsertOryAccessControlPolicy an ORY Access Control Policy.
//
// ```
// PUT /engines/acp/ory/{flavor}/policies HTTP/1.1
// Content-Type: application/json
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#upsertoryaccesscontrolpolicy
func (client *Client) UpsertOryAccessControlPolicy(flavor Flavor, request *UpsertORYAccessPolicyRequest) (*UpsertORYAccessPolicyResponseOK, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err := enc.Encode(request)
	if err != nil {
		return nil, err
	}

	response, err := client.client.Put(client._url+"/engines/acp/ory/"+string(flavor)+"/policies", buf, nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		r := &UpsertORYAccessPolicyResponseOK{}
		err := json.NewDecoder(response.Body).Decode(r)
		if err != nil {
			return nil, err
		}
		return r, nil
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

// ListOryAccessControlPolicy list ORY Access Control Policies.
//
// ```
// GET /engines/acp/ory/{flavor}/policies HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#listoryaccesscontrolpolicies
func (client *Client) ListOryAccessControlPolicy(flavor Flavor, request *ListORYAccessPolicyRequest) (*ListORYAccessPolicyResponseOK, error) {
	s := ""
	if request.Limit > 0 {
		s += fmt.Sprintf("limit=%d", request.Limit)
	}
	if request.Offset > 0 {
		if s != "" {
			s += "&"
		}
		s += fmt.Sprintf("offset=%d", request.Offset)
	}

	if s != "" {
		s = "?" + s
	}

	response, err := client.client.Get(client._url+"/engines/acp/ory/"+string(flavor)+"/policies"+s, nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		r := &ListORYAccessPolicyResponseOK{
			Policies: make([]ORYAccessControlPolicy, 0, 10),
		}
		err := json.NewDecoder(response.Body).Decode(&r.Policies)
		if err != nil {
			return nil, err
		}
		return r, nil
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

// GetOryAccessControlPolicy list ORY Access Control Policies.
//
// ```
// GET /engines/acp/ory/{flavor}/policies/{id} HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#getoryaccesscontrolpolicy
func (client *Client) GetOryAccessControlPolicy(flavor Flavor, id string) (*GetORYAccessPolicyResponseOK, error) {
	response, err := client.client.Get(client._url+"/engines/acp/ory/"+string(flavor)+"/policies/"+id, nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		r := &GetORYAccessPolicyResponseOK{}
		err := json.NewDecoder(response.Body).Decode(&r.Policy)
		if err != nil {
			return nil, err
		}
		return r, nil
	case http.StatusNotFound:
		return nil, ErrNotFound
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

// DeleteOryAccessControlPolicy deletes an ORY Access Control Policy.
//
// ```
// DELETE /engines/acp/ory/{flavor}/policies/{id} HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#deleteoryaccesscontrolpolicy
func (client *Client) DeleteOryAccessControlPolicy(flavor Flavor, id string) error {
	response, err := client.client.Delete(client._url+"/engines/acp/ory/"+string(flavor)+"/policies/"+id, nil)
	if err != nil {
		return err
	}

	switch response.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusInternalServerError:
		r := &ResponseError{}
		dec := json.NewDecoder(response.Body)
		err := dec.Decode(r)
		if err != nil {
			return err
		}
		return r
	default:
		return &UnexpectedResponse{Response: response}
	}
}

// UpsertOryAccessControlRole update or insert a ORY Access Control Role.
//
// Roles group several subjects into one. Rules can be assigned to ORY Access
// Control Policy (OACP) by using the Role ID as subject in the OACP.
//
// ```
// PUT /engines/acp/ory/{flavor}/roles HTTP/1.1
// Content-Type: application/json
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#upsert-an-ory-access-control-policy-role
func (client *Client) UpsertOryAccessControlRole(flavor Flavor, request *UpsertORYAccessRoleRequest) (*UpsertORYAccessRoleResponseOK, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err := enc.Encode(request.Role)
	if err != nil {
		return nil, err
	}

	response, err := client.client.Put(client._url+"/engines/acp/ory/"+string(flavor)+"/roles", buf, nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		r := &UpsertORYAccessRoleResponseOK{}
		err := json.NewDecoder(response.Body).Decode(&r.Role)
		if err != nil {
			return nil, err
		}
		return r, nil
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

// GetOryAccessControlRole return a ORY Access Control Role by ID.
//
// ```
// GET /engines/acp/ory/{flavor}/roles/{id} HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#get-an-ory-access-control-policy-role
func (client *Client) GetOryAccessControlRole(flavor Flavor, id string) (*GetORYAccessRoleResponseOK, error) {
	response, err := client.client.Get(client._url+"/engines/acp/ory/"+string(flavor)+"/roles/"+id, nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		r := &GetORYAccessRoleResponseOK{}
		err := json.NewDecoder(response.Body).Decode(&r.Role)
		if err != nil {
			return nil, err
		}
		return r, nil
	case http.StatusNotFound:
		return nil, ErrNotFound
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

// ListOryAccessControlRole list ORY Access Control Roles.
//
// ```
// GET /engines/acp/ory/{flavor}/roles HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#list-ory-access-control-policy-roles
func (client *Client) ListOryAccessControlRole(flavor Flavor, request *ListORYAccessRoleRequest) (*ListORYAccessRoleResponseOK, error) {
	s := ""
	if request.Limit > 0 {
		s += fmt.Sprintf("limit=%d", request.Limit)
	}
	if request.Offset > 0 {
		if s != "" {
			s += "&"
		}
		s += fmt.Sprintf("offset=%d", request.Offset)
	}

	if s != "" {
		s = "?" + s
	}

	response, err := client.client.Get(client._url+"/engines/acp/ory/"+string(flavor)+"/roles"+s, nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		r := &ListORYAccessRoleResponseOK{
			Roles: make([]ORYAccessControlRole, 0, request.Limit),
		}
		err := json.NewDecoder(response.Body).Decode(&r.Roles)
		if err != nil {
			return nil, err
		}
		return r, nil
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

// DeleteOryAccessControlRole deletes an ORY Access Control Role.
//
// ```
// DELETE /engines/acp/ory/{flavor}/roles/{id} HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#delete-an-ory-access-control-policy-role
func (client *Client) DeleteOryAccessControlRole(flavor Flavor, id string) error {
	response, err := client.client.Delete(client._url+"/engines/acp/ory/"+string(flavor)+"/roles/"+id, nil)
	if err != nil {
		return err
	}

	switch response.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusInternalServerError:
		r := &ResponseError{}
		dec := json.NewDecoder(response.Body)
		err := dec.Decode(r)
		if err != nil {
			return err
		}
		return r
	default:
		return &UnexpectedResponse{Response: response}
	}
}

// AddMembersOryAccessControlRole deletes an ORY Access Control Policy.
//
// ```
// PUT /engines/acp/ory/{flavor}/roles/{id}/members HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#add-a-member-to-an-ory-access-control-policy-role
func (client *Client) AddMembersOryAccessControlRole(flavor Flavor, id string, request *AddMembersORYAccessRoleRequest) (*AddMembersORYAccessRoleResponseOK, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	err := enc.Encode(request)
	if err != nil {
		return nil, err
	}

	response, err := client.client.Put(client._url+"/engines/acp/ory/"+string(flavor)+"/roles/"+id+"/members", buf, nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		r := &AddMembersORYAccessRoleResponseOK{}
		err := json.NewDecoder(response.Body).Decode(&r.Role)
		if err != nil {
			return nil, err
		}
		return r, nil
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

// RemoveMemberOryAccessControlRole removes a member from an ORY Access Control
// Role.
//
// ```
// DELETE /engines/acp/ory/{flavor}/roles/{id}/members/{member} HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#remove-a-member-from-an-ory-access-control-policy-role
func (client *Client) RemoveMemberOryAccessControlRole(flavor Flavor, id, member string) error {
	response, err := client.client.Delete(client._url+"/engines/acp/ory/"+string(flavor)+"/roles/"+id+"/members/"+member, nil)
	if err != nil {
		return err
	}

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusInternalServerError:
		r := &ResponseError{}
		dec := json.NewDecoder(response.Body)
		err := dec.Decode(r)
		if err != nil {
			return err
		}
		return r
	default:
		return &UnexpectedResponse{Response: response}
	}
}

// HealthAlive returns a 200 status code when the HTTP server is up running.
//
// ```
// GET /health/alive HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#check-alive-status
func (client *Client) HealthAlive() (*HealthAliveResponse, error) {
	response, err := client.client.Get(client._url+"/health/alive", nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		r := &HealthAliveResponse{}
		dec := json.NewDecoder(response.Body)
		err := dec.Decode(r)
		if err != nil {
			return nil, err
		}
		return r, nil
	case http.StatusServiceUnavailable:
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

// HealthReadness returns a 200 status code when the HTTP server is up running
// and the environment dependencies (e.g. the database) are responsive as well.
//
// ```
// GET /health/ready HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#check-readiness-status
func (client *Client) HealthReadness() (*HealthReadnessResponse, error) {
	response, err := client.client.Get(client._url+"/health/ready", nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		r := &HealthReadnessResponse{}
		dec := json.NewDecoder(response.Body)
		err := dec.Decode(r)
		if err != nil {
			return nil, err
		}
		return r, nil
	case http.StatusServiceUnavailable:
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

// Version returns the service version typically notated using semantic
// versioning.
//
// ```
// GET /version HTTP/1.1
// Accept: application/json
// ```
//
// See Also https://www.ory.sh/docs/keto/sdk/api#get-service-version
func (client *Client) Version() (*VersionResponse, error) {
	response, err := client.client.Get(client._url+"/version", nil)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		r := &VersionResponse{}
		dec := json.NewDecoder(response.Body)
		err := dec.Decode(r)
		if err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, &UnexpectedResponse{Response: response}
	}
}

func (client *Client) CheckVersion() error {
	response, err := client.Version()
	if err != nil {
		return err
	}
	v := response.Version
	if strings.HasPrefix(v, "v") {
		v = v[1:]
	}
	sv, err := semver.Make(v)
	if err != nil {
		return err
	}
	if clientVersion(sv) {
		return nil
	}
	return errors.Wrap(ErrServerIncompatible, errors.Message("got "+response.Version))
}
