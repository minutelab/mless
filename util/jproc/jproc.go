// Package jproc handle processes that communicate using json on stdin/stdout.
package jproc

import (
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"sync"
)

// Processes represents a process that get requests as json objects streamed
// to stdin (separated by a new line), and for each request, in order it send a reply to stdout
type Processes struct {
	cmd  *exec.Cmd
	done <-chan struct{} // the done channel is closed when the process dies
	err  error           // err return the error from the process

	lock    sync.Mutex // the mutex make sure that only one goroutine read/write request (or close)
	encoder *json.Encoder
	stdout  io.Closer // we read the context through the decoder
	decoder *json.Decoder
}

// Start an new process based on the command line
func Start(cmd *exec.Cmd) (*Processes, error) {
	j := &Processes{cmd: cmd}

	if err := j.start(); err != nil {
		return nil, err
	}
	return j, nil
}

// StartWithStderr an new process based on the command line
// and register a callback to be called for each line of stderr
func StartWithStderr(cmd *exec.Cmd, stderr func(string)) (*Processes, error) {
	stderrCloser, err := StdErrCallback(cmd, stderr)
	if err != nil {
		return nil, err
	}

	j, err := Start(cmd)
	if err != nil {
		stderrCloser.Close()
	}
	return j, err
}

// Close the process (killing it if it not already dead and stopping waiting goroutines)
func (j *Processes) Close() error {
	if j == nil || j.cmd.Process == nil {
		return errors.New("not running")
	}
	// The lock here is not for data access, it is to prevent from
	// killing a process in the middle of precessing a request
	j.lock.Lock()
	defer j.lock.Unlock()
	return j.cmd.Process.Kill()
}

// Send a request and decode back the reply to reply
func (j *Processes) Send(request, reply interface{}) error {
	j.lock.Lock()
	defer j.lock.Unlock()

	if err := j.encoder.Encode(request); err != nil {
		return err
	}

	return j.decoder.Decode(reply)
}

// IsRunning return true if the process can get data
func (j *Processes) IsRunning() bool {
	if j == nil || j.done == nil {
		return false
	}
	select {
	case <-j.done:
		return false
	default:
		return true
	}
}

// Done return a channel is get closed when the process dies
// Note: If the process never started it return nil
func (j *Processes) Done() <-chan struct{} { return j.done }

// Error return the final error code of the process (or nil if doesn't have one)
func (j *Processes) Error() error { return j.err }

// Start the process
func (j *Processes) start() error {
	r, err := j.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	w, err := j.cmd.StdinPipe()
	if err != nil {
		r.Close()
		return err
	}

	if err := j.cmd.Start(); err != nil {
		r.Close()
		w.Close()
		return err
	}

	c := make(chan struct{})
	j.done = c
	j.stdout = r
	j.decoder = json.NewDecoder(r)
	j.encoder = json.NewEncoder(w)

	go func() {
		_, j.err = j.cmd.Process.Wait()
		close(c)
	}()
	return nil
}
