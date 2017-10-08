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

func allocateName(f formation.Function) (string, runtime.Logger) {
	id := atomic.AddUint32(&rid, 1)
	name := fmt.Sprintf("%s-%d", f.FunctionName, id)

	return name, newLogger(name, f.FunctionName, f.Runtime)
}

func newRunner(f formation.Function, env map[string]string, name string, logger runtime.Logger, debug bool) (runtime.Container, error) {
	settings := lambda.StartupRequest{
		Env:     env,
		Handler: f.Handler,
	}

	p, err := runtime.New(f, settings, name, logger, debug)
	if err != nil {
		return nil, err
	}

	logger.ContainerEvent("Started", nil)
	return p, nil
}
