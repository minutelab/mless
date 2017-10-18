package run

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"golang.org/x/sync/syncmap"

	"github.com/inconshreveable/log15"

	"github.com/minutelab/mless/formation"
	"github.com/minutelab/mless/lambda"
	"github.com/minutelab/mless/mlessd/runtime"
	"github.com/minutelab/mless/util/envstr"
)

var (
	rid     uint32
	runners syncmap.Map
)

type runnerKey struct {
	name     string
	runtime  string
	handler  string
	env      string
	debugger interface{}
}

// Get return an invoker suitable for running the function
//
// It attempt to reuse previous runners as long as they have the same defenitions,
// same environment and and same code.
//
// If a sutiable one is not found a new one is returned
func Get(fn formation.Function, env map[string]string) (lambda.Invoker, error) {
	key := runnerKey{
		name:     fn.FunctionName,
		runtime:  fn.Runtime,
		handler:  fn.Handler,
		debugger: fn.Mless.Debugger,
		env:      envstr.Encode(env),
	}

	hash, err := hashDir(fn.Code())
	if err != nil {
		log15.Info("error hashing dir", "dir", fn.Code, "hash", hash, "err", err)
	}

	if old := getRunner(key, hash); old != nil {
		return old, nil
	}

	name, logger := allocateName(fn)

	debug, reservation, err := useDebugger(&fn, name)
	if err != nil {
		key.debugger = nil
		if old := getRunner(key, hash); old != nil {
			log15.Warn("Could not start a deubbger using old one without a deubbger", "reason", err)
			return old, nil
		}
		log15.Warn("Cannot start a deubbger", "reason", err)
	}

	rnr, err := newRunner(fn, env, name, logger, debug)
	if err != nil {
		reservation.Close()
		return nil, err
	}
	go func() {
		<-rnr.Done()
		logger.ContainerEvent("Terminated", rnr.Err())
		if reservation != nil {
			reservation.Close()
		}
	}()

	newRunner := &runner{Container: rnr, hash: hash}
	runners.Store(key, newRunner)
	return newRunner, nil
}

func getRunner(key runnerKey, hash []byte) lambda.Invoker {
	if val, ok := runners.Load(key); ok {
		old := val.(*runner)
		if old.IsRunning() {
			if bytes.Equal(old.hash, hash) {
				return old
			}
			log15.Info("hash is different, need another runner")
			old.Close()
		}
	}

	return nil
}

func useDebugger(fn *formation.Function, container string) (bool, io.Closer, error) {
	if fn.Mless.Debugger == nil {
		return false, nopCloser, nil
	}
	debugger, ok := fn.Mless.Debugger.(runtime.Debugger)
	if !ok {
		// no need for reservation
		return true, nopCloser, nil
	}
	reservation, err := reserveDebugger(debugger.ReservationKey(), container)
	if err != nil {
		return false, nil, err
	}
	return true, reservation, err
}

var (
	reservationLock      sync.Mutex
	debuggerReservations = make(map[string]string) // mapping from resource to owning container
)

func reserveDebugger(resource, container string) (io.Closer, error) {
	reservationLock.Lock()
	defer reservationLock.Unlock()

	if prev, ok := debuggerReservations[resource]; ok {
		return nil, fmt.Errorf("already used by %s", prev)
	}
	log15.Info("Allocation debug reservation", "resource", resource, "container", container)
	debuggerReservations[resource] = container
	return closer(func() error {
		reservationLock.Lock()
		defer reservationLock.Unlock()

		if prev := debuggerReservations[resource]; prev != container {
			return fmt.Errorf("not holding (%s!=%s)", container, prev)
		}
		delete(debuggerReservations, resource)
		log15.Info("Freeing debug reservation", "resource", resource, "container", container)
		return nil
	}), nil
}

type closer func() error

func (c closer) Close() error { return c() }

var nopCloser = closer(func() error { return nil })
