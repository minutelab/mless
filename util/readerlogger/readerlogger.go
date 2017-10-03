package readerlogger

import (
	"io"

	"github.com/inconshreveable/log15"
)

// New create an io.Reader that log copies of what it read to log15 logger
// created by the context
func New(r io.Reader, ctx ...interface{}) io.Reader {
	return &readLogger{
		r:   r,
		log: log15.New(ctx...),
	}
}

type readLogger struct {
	r   io.Reader
	log log15.Logger
}

func (r *readLogger) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if err != nil {
		r.log.Error("Read Error", "err", err)
	} else {
		r.log.Debug("Read", "n", n, "buf", string(p[:n]))
	}
	return
}
