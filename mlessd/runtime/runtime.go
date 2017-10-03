// Package runtime handle the defenitions of the different runtimes.
//
// Currently only python2.7 is supproted, but addition of more should
// affect this package while the rest of the code should not be aware of the differences
package runtime

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/inconshreveable/log15"

	"github.com/minutelab/mless/formation"
)

var (
	baseDir   string
	desktopIP string
)

// Init the runtime with the location of the runtime scripts
func Init(dir, ip string) {
	baseDir = dir
	desktopIP = ip
}

// Type represent the type of of the runtime (pythos2.7, etc.)
type Type interface {
	// CmdLine create a comamnd line that can start the runner
	CmdLine(f formation.Function, env map[string]string, id int) ([]string, error)
}

var runtimes = map[string]Type{
	"python2.7": pythonFacotoy("2.7"),
	"python3.6": pythonFacotoy("3.6"),
}

// Get a runtime Type according to name
// or nil if it does not exist
func Get(tp string) Type {
	rt, _ := runtimes[tp]
	return rt
}

type pythonFacotoy string

func (f pythonFacotoy) CmdLine(fn formation.Function, env map[string]string, id int) ([]string, error) {
	// When the python process load "C" extention module (for example go programs)
	// It seem that the setting to the environment in the process does not pass to them
	// So instead of relying on the environment transmitted to the runtime process
	// we store it in a file in the container and read it BEFORE the process start
	// TODO: we need to clean this file once the container start
	envfile, err := ioutil.TempFile("", fmt.Sprintf("env-%d.", id))
	if err != nil {
		return nil, err
	}
	for k, v := range env {
		fmt.Fprintf(envfile, "export %s=%s\n", k, v)
	}
	envfile.Close()

	cmdline := []string{
		path.Join(baseDir, "python/python.mlab"),
		"-ver", string(f),
		"-name", fmt.Sprintf("%s-%d", fn.FunctionName, id),
		"-dir", fn.Code(),
		"-envfile", envfile.Name(),
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
