package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/minutelab/mless/lambda"
	"github.com/minutelab/mless/util/jproc"
)

// stdiocont is a runtime container that communication with the container over stdin/stdout
// (based on jproc)
type stdiocont struct {
	proc *jproc.Process
}

func (s *stdiocont) Done() <-chan struct{} { return s.proc.Done() }
func (s *stdiocont) Err() error            { return s.proc.Error() }
func (s *stdiocont) Close() error          { return s.proc.Close() }

func (s *stdiocont) Invoke(event interface{}, context lambda.Context, deadline time.Time) (json.RawMessage, error) {
	request := lambda.InvokeRequest{
		Context:  context,
		Event:    event,
		Deadline: deadline.UnixNano() / 1000000, // convert from nano to mili
	}
	var reply lambda.InvokeReply

	if err := s.proc.Send(request, &reply); err != nil {
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

func newStdiocont(cmdline []string, settings lambda.StartupRequest, name string, id string) (Container, error) {
	cmd := exec.Command(cmdline[0], cmdline[1:]...)

	p, err := jproc.StartWithStderr(cmd, func(s string) { fmt.Fprintf(os.Stderr, "rid-%s: %s\n", id, s) })
	if err != nil {
		return nil, err
	}

	var response lambda.StartupResponse
	if err := p.Send(settings, &response); err != nil {
		p.Close()
		return nil, err
	}
	if !response.OK {
		p.Close()
		return nil, errors.New("runner initialization failed")
	}

	return &stdiocont{p}, nil
}
