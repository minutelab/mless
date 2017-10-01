package run

import (
	"bytes"

	"golang.org/x/sync/syncmap"

	"github.com/inconshreveable/log15"

	"github.com/minutelab/mless/formation"
	"github.com/minutelab/mless/lambda"
	"github.com/minutelab/mless/util/envstr"
)

var (
	rid     uint32
	runners syncmap.Map
)

type runnerKey struct {
	name    string
	runtime string
	handler string
	mless   formation.Mless
	env     string
}

// Get return an invoker suitable for running the function
//
// It attempt to reuse previous runners as long as they have the same defenitions,
// same environment and and same code.
//
// If a sutiable one is not found a new one is returned
func Get(fn formation.Function, env map[string]string) (lambda.Invoker, error) {
	key := runnerKey{
		name:    fn.FunctionName,
		runtime: fn.Runtime,
		handler: fn.Handler,
		mless:   fn.Mless,
		env:     envstr.Encode(env),
	}

	hash, err := hashDir(fn.Code())
	if err != nil {
		log15.Info("error hashing dir", "dir", fn.Code, "hash", hash, "err", err)
	}

	if val, ok := runners.Load(key); ok {
		old := val.(*runner)
		if old.IsRunning() {
			if bytes.Equal(old.hash, hash) {
				return old, nil
			}
			log15.Info("hash is different, need another runner")
			old.Close()
		}
	}

	rnr, err := newRunner(fn, env)
	if err != nil {
		return nil, err
	}
	newRunner := &runner{Container: rnr, hash: hash}
	runners.Store(key, newRunner)
	return newRunner, nil
}
