// Package runtime handle the defenitions of the different runtimes.
//
// Currently only python2.7 is supproted, but addition of more should
// affect this package while the rest of the code should not be aware of the differences
package runtime

import (
	"fmt"
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
	CmdLine(f formation.Function, id int) ([]string, error)
}

var runtimes = map[string]Type{
	"python2.7": python27Facotoy{},
}

// Get a runtime Type according to name
// or nil if it does not exist
func Get(tp string) Type {
	rt, _ := runtimes[tp]
	return rt
}

type python27Facotoy struct{}

func (f python27Facotoy) CmdLine(fn formation.Function, id int) ([]string, error) {
	cmdline := []string{
		path.Join(baseDir, "python2.7/python2.7.mlab"),
		"-name", fmt.Sprintf("%s-%d", fn.FunctionName, id),
		"-dir", fn.Code(),
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
