package runtime

import (
	"fmt"
	"os"
	"path"
	"strings"

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

func newNode610(fn formation.Function, settings lambda.StartupRequest, name string, logger Logger, debug bool) (Container, error) {
	envfile, err := writeEnvFile(settings, name)
	if err != nil {
		return nil, err
	}
	defer os.Remove(envfile)

	cmdline, err := node610CmdLine(fn, name, envfile, debug)
	if err != nil {
		return nil, err
	}

	return newNodeCont(cmdline, settings, name, logger)
}

func node610CmdLine(fn formation.Function, name string, envfile string, debug bool) ([]string, error) {
	cmdline := []string{
		path.Join(baseDir, "nodejs6.10/nodejs6.10.mlab"),
		"-name", name,
		"-dir", fn.Code(),
		"-envfile", envfile,
	}

	if debug && fn.Mless.Debugger != nil {
		cmdline = append(cmdline,
			"-debugger", string(fn.Mless.Debugger.(nodeDebugger)),
			"-dport", "5858",
		)
	}
	return cmdline, nil
}

func parseNodeDebugger(fn formation.Function) (Debugger, error) {
	switch dbg := fn.Mless.Debugger.(type) {
	case bool:
		if dbg {
			return nodeDebugger("legacy"), nil
		}
		return nil, nil
	case string:
		dbg = strings.ToLower(dbg)
		if dbg == "legacy" || dbg == "inspector" {
			return nodeDebugger(dbg), nil
		}
		return nil, fmt.Errorf("unsupported nodejs debugger: %s", dbg)
	default:
		return nil, fmt.Errorf("unsupported nodejs debugger defenition: %v", dbg)
	}
}

type nodeDebugger string

func (p nodeDebugger) ReservationKey() string { return "nodejs" }
