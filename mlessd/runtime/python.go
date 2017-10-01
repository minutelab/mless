package runtime

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/minutelab/mless/formation"
	"github.com/minutelab/mless/lambda"
)

func newPython(ver string, fn formation.Function, settings lambda.StartupRequest, logger log15.Logger, id string) (Container, error) {
	envfile, err := writeEnvFile(settings, id)
	if err != nil {
		return nil, err
	}
	defer os.Remove(envfile)

	cmdline, err := pythonCmdLine(ver, fn, id, envfile)
	if err != nil {
		return nil, err
	}

	return newStdiocont(cmdline, settings, fn.FunctionName, id)
}

func pythonCmdLine(ver string, fn formation.Function, id string, envfile string) ([]string, error) {
	cmdline := []string{
		path.Join(baseDir, "python/python.mlab"),
		"-ver", ver,
		"-name", fmt.Sprintf("%s-%s", fn.FunctionName, id),
		"-dir", fn.Code(),
		"-envfile", envfile,
	}

	switch strings.ToLower(fn.Mless.Debugger) {
	case "pydevd":
		cmdline = append(cmdline,
			"-debugger", "pydevd",
			"-dhost", desktopIP,
			"-desktop", fn.Desktop(),
		)
	case "":
		// nothing
	default:
		log15.Warn("Unknown python 2.7 debugger", "func", fn.FunctionName, "debugger", fn.Mless.Debugger)
	}
	return cmdline, nil
}
