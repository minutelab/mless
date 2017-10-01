package run

import (
	"strconv"
	"sync/atomic"

	"github.com/inconshreveable/log15"

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
	logger := log15.New("rid", id)

	settings := lambda.StartupRequest{
		Env:     env,
		Handler: f.Handler,
	}

	logger.Info("Starting runner", "function", f.FunctionName, "runtime", f.Runtime)

	p, err := runtime.New(f, settings, logger, strconv.Itoa(int(id)))
	if err != nil {
		return nil, err
	}

	logger.Info("Started")
	go func() {
		<-p.Done()
		logger.Info("Terminated", "err", p.Err())
	}()
	return p, nil
}
