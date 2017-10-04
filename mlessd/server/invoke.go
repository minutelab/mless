package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/minutelab/mless/formation"
	"github.com/minutelab/mless/lambda"
	"github.com/minutelab/mless/mlessd/run"
	"github.com/minutelab/mless/util/ldebug"
)

// invokeHandler handle remote invocation of serverless function through the proxy
func invokeHandler(r *http.Request) (interface{}, error) {
	// Parse request
	var request lambda.ProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(time.Duration(request.Remaining) * time.Millisecond)

	fname := request.Env["AWS_LAMBDA_FUNCTION_NAME"]
	if fname == "" {
		return nil, errors.New("no function name")
	}

	functions.Refresh()

	fn := functions.Get(fname)
	if fn == nil {
		return nil, errors.New("no such function")
	}

	// Prepare the environment
	prepareEnv(request.Env, fn)

	fmt.Println("Processed request:", ldebug.DumpJSON(request))

	runner, err := run.Get(*fn, request.Env)
	if err != nil {
		log15.Error("Failed getting runner", "err", err)
		return nil, err
	}

	return runner.Invoke(request.Event, request.Context, deadline)
}

func prepareEnv(env map[string]string, fn *formation.Function) {
	// We will first clean the environment that we got from the proxy
	var cleanEnv = []string{
		// What to do with XRAY?
		// Its seems that the XRAY (usually?) point to private address that
		// we cannot get to, it is probably safer for now to clean them
		"_X_AMZN_TRACE_ID",
		"AWS_XRAY_DAEMON_ADDRESS",
		"_AWS_XRAY_DAEMON_ADDRESS",
		"_AWS_XRAY_DAEMON_PORT",
		// Handler may be different
		"_HANDLER",
		// The exution environment may be different than the proxy
		"AWS_EXECUTION_ENV",
		"PATH",
		"PYTHONPATH",
		"LD_LIBRARY_PATH",
	}
	for _, k := range cleanEnv {
		delete(env, k)
	}

	// now let's merge environment defined in the formation
	if fn.Environment != nil {
		for k, v := range fn.Environment.Variables {
			env[k] = v
		}
	}
}
