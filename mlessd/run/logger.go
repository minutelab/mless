package run

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/minutelab/mless/lambda"
	"github.com/minutelab/mless/mlessd/runtime"
)

type logger struct {
	name string
}

func newLogger(name, fname, runtime string) runtime.Logger {
	l := logger{name: name}
	l.printf(aurora.BlueFg, "Starting container %s:%s", fname, runtime)
	return l
}

func (l logger) StdErr(line string) {
	l.print(aurora.GrayFg, line)
}

func (l logger) ContainerEvent(event string, err error) {
	if err == nil {
		l.print(aurora.BlueFg, event)
	} else {
		l.printf(aurora.RedFg, "%s : %s", event, err)
	}
}

func (l logger) FunctionResult(res *lambda.InvokeReply) {
	color := aurora.GreenFg
	if res.Errors {
		color = aurora.RedFg
	}

	l.printf(color,
		"REPORT RequestId: %s Duration: %.02f ms Billed Duration: %d ms Memory Size: %s MB Max Memory Used: %d MB",
		res.InvokeID,
		res.Billing.Duraion,
		100*int((1+math.Floor(res.Billing.Duraion/100))),
		res.Billing.Memory,
		res.Billing.Used)

	if res.Errors {
		l.printf(color, "%s (%s)", string(res.Result), res.ErrorType)
		return
	}
	l.print(color, string(res.Result))
}

func (l logger) printf(color aurora.Color, format string, a ...interface{}) {
	l.print(color, fmt.Sprintf(format, a...))
}

func (l logger) print(color aurora.Color, line string) {
	s := fmt.Sprintf("%s %s: %s\n", time.Now().Format("01-02 15:04:05.000"), l.name, line)
	fmt.Fprint(os.Stderr, aurora.Colorize(s, color))
}
