package run

import (
	"fmt"
	"sync/atomic"

	"github.com/minutelab/mless/formation"
	"github.com/minutelab/mless/lambda"
	"github.com/minutelab/mless/mlessd/runtime"
)

type runner struct {
	runtime.Container
	hash []byte
}

// IsRunning return true if the process can get data
func (r *runner) IsRunning() bool {
	if r == nil || r.Done() == nil {
		return false
	}
	select {
	case <-r.Done():
		return false
	default:
		return true
	}
}

func newRunner(f formation.Function, env map[string]string) (runtime.Container, error) {
	id := atomic.AddUint32(&rid, 1)
	name := fmt.Sprintf("%s-%d", f.FunctionName, id)
	logger := newLogger(name, f.FunctionName, f.Runtime)

	settings := lambda.StartupRequest{
		Env:     env,
		Handler: f.Handler,
	}

	p, err := runtime.New(f, settings, name, logger)
	if err != nil {
		return nil, err
	}

	logger.ContainerEvent("Started", nil)
	go func() {
		<-p.Done()
		logger.ContainerEvent("Terminated", p.Err())
	}()
	return p, nil
}
