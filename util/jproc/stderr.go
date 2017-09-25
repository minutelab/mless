package jproc

import (
	"bufio"
	"io"
	"os/exec"
)

// StdErrCallback provide an "easy" way to read line from a subprocess stdrror stream.
//
// It should be called on *exec.Cmd before calling Run/Start
// If it doesn't return an error and for some reason Wait is not called for the process,
// the Close of the closer should be called.
func StdErrCallback(cmd *exec.Cmd, cb func(string)) (io.Closer, error) {
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			cb(scanner.Text())
		}
	}()

	return stderr, nil
}
