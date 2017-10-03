package runtime

import (
	"fmt"
	"os"
	"path"

	"github.com/inconshreveable/log15"
	"github.com/minutelab/mless/formation"
	"github.com/minutelab/mless/lambda"
)

// With nodejs we have a limitations on the communication with the runtime container:
// When the lambda function is actually running we shouldn't have any event waiting in the node event loop.
// This is because that as long as there are events, the lambda function won't finish running
//
// When sending event information over stdin, I wasn't able to completly remove events from the event loop,
// so the solution was to send the events over http. But we cannot send back the events on http,
// because then we have a connection open while the lambda function is running.
// So we send an empty reply immediatly from node, and on the go side we close the connection.
//
// Then when we have a reply we send it on stdout. So we end up with a strangly mixed protocol, but it work

func newNode610(fn formation.Function, settings lambda.StartupRequest, logger log15.Logger, id string) (Container, error) {
	envfile, err := writeEnvFile(settings, id)
	if err != nil {
		return nil, err
	}
	defer os.Remove(envfile)

	name := fmt.Sprintf("%s-%s", fn.FunctionName, id)
	cmdline, err := node610CmdLine(fn, name, envfile)
	if err != nil {
		return nil, err
	}

	return newNodeCont(cmdline, settings, name, logger, id)
}

func node610CmdLine(fn formation.Function, name string, envfile string) ([]string, error) {
	cmdline := []string{
		path.Join(baseDir, "nodejs6.10/nodejs6.10.mlab"),
		"-name", name,
		"-dir", fn.Code(),
		"-envfile", envfile,
	}
	return cmdline, nil
}
