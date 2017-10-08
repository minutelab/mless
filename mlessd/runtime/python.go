package runtime

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/minutelab/mless/formation"
	"github.com/minutelab/mless/lambda"
)

func newPython(ver string, fn formation.Function, settings lambda.StartupRequest, name string, logger Logger, debug bool) (Container, error) {
	envfile, err := writeEnvFile(settings, name)
	if err != nil {
		return nil, err
	}
	defer os.Remove(envfile)

	cmdline, err := pythonCmdLine(ver, fn, name, envfile, debug)
	if err != nil {
		return nil, err
	}

	return newStdiocont(cmdline, settings, logger)
}

func pythonCmdLine(ver string, fn formation.Function, name string, envfile string, debug bool) ([]string, error) {
	cmdline := []string{
		path.Join(baseDir, "python/python.mlab"),
		"-ver", ver,
		"-name", name,
		"-dir", fn.Code(),
		"-envfile", envfile,
	}

	if debug && fn.Mless.Debugger != nil {
		cmdline = append(cmdline,
			"-debugger", string(fn.Mless.Debugger.(pythonDebugger)),
			"-dhost", desktopIP,
			"-desktop", fn.Desktop(),
		)
	}
	return cmdline, nil
}

func parsePythonDebugger(fn formation.Function) (Debugger, error) {
	tp, ok := fn.Mless.Debugger.(string)
	if !ok {
		return nil, errors.New("python debugger defenition need to be a string")
	}
	tp = strings.ToLower(tp)
	switch tp {
	case "":
		return nil, nil
	case "pydevd", "pydev":
		return pythonDebugger("pydevd"), nil
	default:
		return nil, fmt.Errorf("unsupported python debugger: %s", tp)
	}
}

type pythonDebugger string

func (p pythonDebugger) ReservationKey() string { return "python" }
