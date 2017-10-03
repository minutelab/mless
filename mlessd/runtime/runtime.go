// Package runtime handle the defenitions of the different runtimes.
//
// Currently only python2.7 is supproted, but addition of more should
// affect this package while the rest of the code should not be aware of the differences
package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/inconshreveable/log15"
	"github.com/minutelab/mless/formation"
	"github.com/minutelab/mless/lambda"
)

var (
	baseDir   string
	desktopIP string
)

// Init the runtime with the location of the runtime scripts
func Init(dir, ip string) {
	baseDir = dir
	desktopIP = ip
}

// Container is a specific runtime container
type Container interface {
	lambda.Invoker
	// Closes (stop) the container
	Close() error
	// A channel that is closed when the container exit
	Done() <-chan struct{}
	// Err is valid only after the container exit (the Done channel is closed)
	// And contain the exit error of the container
	Err() error
}

// New create a new runner
func New(fn formation.Function, settings lambda.StartupRequest, logger log15.Logger, id string) (Container, error) {
	switch fn.Runtime {
	case "python2.7":
		return newPython("2.7", fn, settings, logger, id)
	case "python3.6":
		return newPython("3.6", fn, settings, logger, id)
	case "nodejs6.10":
		return newNode610(fn, settings, logger, id)
	}

	return nil, fmt.Errorf("no such runtime: %s", fn.Runtime)
}

func writeEnvFile(settings lambda.StartupRequest, id string) (string, error) {
	envfile, err := ioutil.TempFile("", fmt.Sprintf("env-%s.", id))
	if err != nil {
		return "", err
	}
	defer envfile.Close()
	for k, v := range settings.Env {
		if err := printEnvItem(envfile, k, v); err != nil {
			return "", err
		}
	}
	if err := printEnvItem(envfile, "_HANDLER", settings.Handler); err != nil {
		return "", err
	}
	return envfile.Name(), nil
}

func printEnvItem(o io.Writer, k, v string) error {
	_, err := fmt.Fprintf(o, "export %s=%s\n", k, v)
	return err
}

// runtimeError is the type of error returned by the lambda runtime
type runtimeError struct {
	Type  string
	Value string
}

func (r runtimeError) Error() string { return r.Value }

func processReply(reply lambda.InvokeReply, requestID string) (json.RawMessage, error) {
	if reply.InvokeID != requestID {
		return nil, fmt.Errorf("response ID %s does not match request: %s", reply.InvokeID, requestID)
	}

	if reply.Errors {
		return nil, runtimeError{Type: reply.ErrorType, Value: string(reply.Result)}
	}
	return reply.Result, nil
}
