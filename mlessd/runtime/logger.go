package runtime

import "github.com/minutelab/mless/lambda"

// Logger is the type used to log function invocation
type Logger interface {
	// StdErr log lines sent on the container runtime, which inlcude logging
	// of the function (and its framework processing)
	StdErr(line string)
	// ContainerEvent is used to log things related to the container management
	ContainerEvent(event string, err error)
	// FunctionResult log the result of invocation
	FunctionResult(res *lambda.InvokeReply)
}
