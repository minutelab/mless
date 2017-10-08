package runtime

import (
	"fmt"

	"github.com/minutelab/mless/formation"
)

type Debugger interface {
	ReservationKey() string
}

func parseDebugger(fn formation.Function) (Debugger, error) {
	if fn.Mless.Debugger == nil {
		return nil, nil
	}
	switch fn.Runtime {
	case "python2.7", "python3.6":
		return parsePythonDebugger(fn)

	case "nodejs6.10":
		return parseNodeDebugger(fn)

	default:
		return nil, fmt.Errorf("No debugger implementation for %s", fn.Runtime)
	}
}
