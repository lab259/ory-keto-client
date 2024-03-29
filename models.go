package ketoclient

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Effect string

const (
	Allow Effect = "allow"
	Deny  Effect = "deny"
)

type ORYAccessControlPolicy struct {
	Actions     []string    `json:"actions"`
	Conditions  interface{} `json:"conditions"`
	Description string      `json:"description"`
	Effect      Effect      `json:"effect"`
	ID          string      `json:"id"`
	Resources   []string    `json:"resources"`
	Subjects    []string    `json:"subjects"`
}

type ORYAccessControlRole struct {
	ID      string   `json:"id"`
	Members []string `json:"members"`
}

// ResponseError is the default error format for the Keto service.
type ResponseError struct {
	Code    int64           `json:"code"`
	Details json.RawMessage `json:"details"`
	Message string          `json:"message,omitempty"`
	Reason  string          `json:"reason,omitempty"`
	Request string          `json:"request,omitempty"`
	Status  string          `json:"status,omitempty"`
}

func (err *ResponseError) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("[%d]", err.Code))
	if err.Message != "" {
		buf.WriteString(": ")
		buf.WriteString(err.Message)
	}
	if len(err.Details) > 0 {
		buf.WriteByte(' ')
		buf.Write(err.Details)
	}
	if err.Reason != "" {
		buf.WriteString(": ")
		buf.WriteString(err.Reason)
	}
	return buf.String()
}

/**
 * POST /engines/acp/ory/{flavor}/allowed HTTP/1.1
 * Content-Type: application/json
 * Accept: application/json
 */

type AllowedORYAccessControlPolicyRequest struct {
	Action   string      `json:"action"`
	Context  interface{} `json:"context"`
	Resource string      `json:"resource"`
	Subject  string      `json:"subject"`
}

type AllowedORYAccessControlPolicyResponse struct {
	Allowed bool `json:"allowed"`
}

/**
 * AcpUpsertORYAccessPolicy
 *
 * PUT /engines/acp/ory/{flavor}/policies HTTP/1.1
 * Content-Type: application/json
 * Accept: application/json
 */

type UpsertORYAccessPolicyRequest struct {
	ORYAccessControlPolicy
}

type UpsertORYAccessPolicyResponseOK struct {
	*ORYAccessControlPolicy
}

/**
 * GET /engines/acp/ory/{flavor}/policies HTTP/1.1
 * Accept: application/json
 */

type ListORYAccessPolicyRequest struct {
	Limit  int64
	Offset int64
}

type ListORYAccessPolicyResponseOK struct {
	Policies []ORYAccessControlPolicy
}

/**
 * GET /engines/acp/ory/{flavor}/policies/{id} HTTP/1.1
 * Accept: application/json
 */

type GetORYAccessPolicyResponseOK struct {
	Policy ORYAccessControlPolicy
}

/**
 * PUT /engines/acp/ory/{flavor}/roles HTTP/1.1
 * Content-Type: application/json
 * Accept: application/json
 */

type UpsertORYAccessRoleRequest struct {
	Role ORYAccessControlRole
}

type UpsertORYAccessRoleResponseOK struct {
	Role ORYAccessControlRole
}

/**
 * GET /engines/acp/ory/{flavor}/roles/{id} HTTP/1.1
 * Content-Type: application/json
 * Accept: application/json
 */

type GetORYAccessRoleResponseOK struct {
	Role ORYAccessControlRole
}

/**
 * GET /engines/acp/ory/{flavor}/roles HTTP/1.1
 * Accept: application/json
 */

type ListORYAccessRoleRequest struct {
	Limit  int64
	Offset int64
}

type ListORYAccessRoleResponseOK struct {
	Roles []ORYAccessControlRole
}

/**
 * GET /engines/acp/ory/{flavor}/roles HTTP/1.1
 * Accept: application/json
 */

type AddMembersORYAccessRoleRequest struct {
	Members []string `json:"members"`
}

type AddMembersORYAccessRoleResponseOK struct {
	Role ORYAccessControlRole
}

/**
 * GET /health/alive HTTP/1.1
 * Accept: application/json
 */

type HealthAliveResponse struct {
	Status string `json:"status"`
}

/**
 * GET /health/readness HTTP/1.1
 * Accept: application/json
 */

type HealthReadnessResponse struct {
	Status string `json:"status"`
}

/**
 * GET /version HTTP/1.1
 * Accept: application/json
 */

type VersionResponse struct {
	Version string `json:"version"`
}
