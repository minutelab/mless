package runtime

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/minutelab/mless/lambda"
	"github.com/minutelab/mless/util/jproc"
)

// nodecont is a runtime container that communication with the node container
// It send request over http and recieve them over stdout...
type nodecont struct {
	cmd     *exec.Cmd
	logger  Logger
	url     string
	decoder *json.Decoder
	done    <-chan struct{}
	err     error
	lock    sync.Mutex // serialization lock
}

var (
	// we want our own transport to control over idle conenction (we don't want them)
	transport = &http.Transport{
		MaxIdleConns:       1,
		IdleConnTimeout:    time.Millisecond,
		DisableCompression: true,
	}

	client = &http.Client{
		Transport: transport,
	}
)

func newNodeCont(cmdline []string, settings lambda.StartupRequest, name string, logger Logger) (Container, error) {
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.SysProcAttr = setPdeathsig(nil, syscall.SIGKILL)

	stderrCloser, err := jproc.StdErrCallback(cmd, logger.StdErr)
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stderrCloser.Close()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	logger.ContainerEvent("Started container waiting for ack", nil)
	decoder := json.NewDecoder(stdout)
	// decoder := json.NewDecoder(readerlogger.New(stdout, "stdout", name))

	var res lambda.StartupResponse
	switch err := decoder.Decode(&res); {
	case err != nil:
		return nil, fmt.Errorf("failed decoding startup reply: %s", err)
	case !res.OK:
		return nil, errors.New("got wrong startup result")
	}

	cont := &nodecont{
		cmd:     cmd,
		logger:  logger,
		decoder: decoder,
		url:     fmt.Sprintf("http://%s:%d/", name, 8999),
	}

	cont.done = cont.waiter()

	return cont, nil
}

func (h *nodecont) Done() <-chan struct{} { return h.done }
func (h *nodecont) Err() error            { return h.err }

func (h *nodecont) Close() error {
	if h == nil || h.cmd.Process == nil {
		return errors.New("not running")
	}
	// The lock here is not for data access, it is to prevent from
	// killing a process in the middle of precessing a request
	h.lock.Lock()
	defer h.lock.Unlock()
	return h.cmd.Process.Kill()
}

func (h *nodecont) waiter() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		state, err := h.cmd.Process.Wait()
		switch {
		case err != nil:
			h.err = err
		case !state.Success():
			h.err = errors.New(state.String())
		}
		close(done)
	}()
	return done
}

func (h *nodecont) Invoke(event interface{}, context lambda.Context, deadline time.Time) (json.RawMessage, error) {
	request := lambda.InvokeRequest{
		Context:  context,
		Event:    event,
		Deadline: deadline.UnixNano() / 1000000, // convert from nano to mili
	}

	reqBuf, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	// Lock is to serialize requests
	h.lock.Lock()
	defer h.lock.Unlock()
	defer transport.CloseIdleConnections()

	response, err := client.Post(h.url, "application/json", bytes.NewReader(reqBuf))
	if err != nil {
		return nil, err
	}
	response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("Http error %s(%d): %s", response.Status, response.StatusCode, err)
		}
		return nil, fmt.Errorf("Http error %s(%d): %q", response.Status, response.StatusCode, string(body))
	}
	http.DefaultTransport.(*http.Transport).CloseIdleConnections()

	var reply lambda.InvokeReply
	if err := h.decoder.Decode(&reply); err != nil {
		return nil, err
	}
	return processReply(reply, context.RequestID, h.logger)
}
