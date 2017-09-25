package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/minutelab/mless/formation"
	"github.com/minutelab/mless/lambda"
	"github.com/minutelab/mless/mlessd/runtime"
	"github.com/minutelab/mless/util/jproc"
)

type Invoker interface {
	Invoke(event interface{}, context lambda.Context, deadline time.Time) (json.RawMessage, error)
}

// RuntimeError is the type of error returned by the runtime
type RuntimeError struct {
	Type  string
	Value string
}

func (r RuntimeError) Error() string { return r.Value }

type runner struct {
	proc *jproc.Processes
	hash []byte
}

func (r *runner) Invoke(event interface{}, context lambda.Context, deadline time.Time) (json.RawMessage, error) {
	request := lambda.InvokeRequest{
		Context:  context,
		Event:    event,
		Deadline: deadline.UnixNano() / 1000000, // convert from nano to mili
	}
	var reply lambda.InvokeReply

	if err := r.proc.Send(request, &reply); err != nil {
		return nil, err
	}

	if reply.InvokeID != context.RequestID {
		return nil, fmt.Errorf("response ID %s does not match request: %s", reply.InvokeID, context.RequestID)
	}

	if reply.Errors {
		return nil, RuntimeError{Type: reply.ErrorType, Value: string(reply.Result)}
	}
	return reply.Result, nil
}

func newRunner(f formation.Function, env map[string]string) (*jproc.Processes, error) {
	id := atomic.AddUint32(&rid, 1)
	logger := log15.New("rid", id)
	logger.Info("Starting runner", "function", f.FunctionName)

	rtime := runtime.Get(f.Runtime)
	if rtime == nil {
		return nil, fmt.Errorf("no runtime for %s", f.Runtime)
	}

	cmdline, err := rtime.CmdLine(f, env, int(id))
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(cmdline[0], cmdline[1:]...)

	p, err := jproc.StartWithStderr(cmd, func(s string) { fmt.Fprintf(os.Stderr, "rid-%d: %s\n", id, s) })

	if err != nil {
		return nil, err
	}

	request := lambda.StartupRequest{
		Handler: f.Handler,
		Env:     env,
	}
	var response lambda.StartupResponse
	if err := p.Send(request, &response); err != nil {
		p.Close()
		return nil, err
	}
	if !response.OK {
		p.Close()
		return nil, errors.New("runner initialization failed")
	}

	logger.Info("Started")
	go func() {
		<-p.Done()
		logger.Info("Terminated", "err", p.Error())
	}()
	return p, nil
}
