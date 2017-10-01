// Package lambda contain structure defentition for lambda information.
//
// It include structures that are passed between the proxy and mlessd, and between mlessd and the runtime containers.
package lambda

import (
	"encoding/json"
	"time"
)

// Invoker is the interface for a container running lambda function
// The Invoker is responsbile for serialization of the requests.
type Invoker interface {
	Invoke(event interface{}, context Context, deadline time.Time) (json.RawMessage, error)
}

// ProxyRequest is the request sent from the proxy to mlabd
type ProxyRequest struct {
	Context   Context           `json:"context"`
	Remaining int               `json:"remaining"` // remaining time in millis
	Env       map[string]string `json:"env"`
	Event     interface{}       `json:"event"`
}

// StartupRequest include the initial request sent to the a runner
type StartupRequest struct {
	Handler string            `json:"handler"`
	Env     map[string]string `json:"env"`
}

// StartupResponse is the reply sent by the runtime to a (succesful startup)
type StartupResponse struct {
	OK bool `json:"ok"`
}

// InvokeRequest is the type of message sent to the runner to start processing of a specific event
type InvokeRequest struct {
	Context  Context     `json:"context"`
	Event    interface{} `json:"event"`
	Deadline int64       `json:"deadline"` // number of milliseconds since January 1, 1970 UTC.
}

// InvokeReply is the type of message sent to the runner to start processing of a specific event
type InvokeReply struct {
	Result    json.RawMessage `json:"result"`
	ErrorType string          `json:"errortype"`
	InvokeID  string          `json:"invokeid"`
	Errors    bool            `json:"errors,omitempty"`
}

// Context include the context data that is passed to the client
// Note: It does not include all the data that is stored in the context,
// some fields are duplicated in the environment and we take the value from there.
type Context struct {
	RequestID string         `json:"aws_request_id"`
	Client    *ClientContext `json:"client_context"`
	Identity  *Identity      `json:"identity"`
	ARN       string         `json:"invoked_function_arn"`
}

type Identity struct {
	ID     string `json:"cognito_identity_id"`
	PoolID string `json:"cognito_identity_pool_id"`
}

type ClientContext struct {
	Client ClientClient      `json:"client"`
	Custom map[string]string `json:"custom"`
	Env    map[string]string `json:"env"`
}

type ClientClient struct {
	InstallationID string `json:"installation_id"`
	AppTitle       string `json:"app_title"`
	AppVersionName string `json:"app_version_name"`
	AppVersionCode string `json:"app_version_code"`
	AppPackageName string `json:"app_package_name"`
}
