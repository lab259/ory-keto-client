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

type AcpAllowedRequest struct {
	Action   string      `json:"action"`
	Context  interface{} `json:"context"`
	Resource string      `json:"resource"`
	Subject  string      `json:"subject"`
}

type AcpAllowedResponse struct {
	Allowed bool `json:"allowed"`
}

/**
 * AcpPutPolicies
 *
 * PUT /engines/acp/ory/{flavor}/policies HTTP/1.1
 * Content-Type: application/json
 * Accept: application/json
 */

type AcpUpsertORYAccessPolicyRequest struct {
	ORYAccessControlPolicy
}

type AcpUpsertORYAccessPolicyResponseOK struct {
	*ORYAccessControlPolicy
}
