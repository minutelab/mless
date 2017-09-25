package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/alecthomas/kingpin"

	"github.com/minutelab/mless/formation"
	"github.com/minutelab/mless/mlessd/runtime"
	"github.com/minutelab/mless/mlessd/server"
)

var (
	Version   string
	BuildDate string
)

var ctx struct {
	template  string
	srcdir    string
	runtime   string
	desktopIP string
}

func main() {
	app := kingpin.New("mless", "run serverless code in mlab")
	app.DefaultEnvars()
	app.HelpFlag.Short('h')
	app.UsageTemplate(kingpin.CompactUsageTemplate)

	app.Command("version", "Show application version").Action(showVersion)

	daemon := app.Command("start", "start mlessd daemon").Action(mlessd)
	daemon.Arg("tempalte", "template file").Default("/sam/template.yaml").ExistingFileVar(&ctx.template)
	daemon.Flag("src", "location sources on desktop").StringVar(&ctx.srcdir)
	daemon.Flag("desktop", "IP of desktop (for back c onnection)").StringVar(&ctx.desktopIP)
	daemon.Flag("runtime", "location of lambda runtime dircotry").Default("/usr/local/mless/runtime").ExistingDirVar(&ctx.runtime)

	if _, err := app.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func showVersion(_ *kingpin.ParseContext) error {
	if Version == "" {
		fmt.Println("develop")
	} else {
		fmt.Printf("%s (%s)\n", Version, BuildDate)
	}
	return nil
}

func mlessd(_ *kingpin.ParseContext) error {
	runtime.Init(ctx.runtime, ctx.desktopIP)

	functions, err := formation.New(ctx.template, path.Clean(ctx.srcdir), 2*time.Second)
	if err != nil {
		return err
	}

	return server.Start(8000, functions)
}
